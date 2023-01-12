const AWS = require("aws-sdk");
const fs = require("fs");
const JSZip = require("jszip");
const dotenv = require("dotenv");

dotenv.config();

const s3 = new AWS.S3({
  endpoint: "https://nyc3.digitaloceanspaces.com",
  region: "us-east-1",
  credentials: {
    accessKeyId: process.env.S3_ACCESS_KEY,
    secretAccessKey: process.env.S3_SECRET_ACCESS_KEY,
  },
});

console.log("Downloading...");

function filterDemoByTier(data, tierName) {
  return data.Contents.filter((item) => {
    return item.Key.includes(tierName);
  });
}

async function processTier(tierName, bucketPrefix) {
  console.log(`Processing tier ${tierName}`);

  return new Promise((res) => {
    s3.listObjectsV2(
      {
        Bucket: "cscdemos",
        Prefix: bucketPrefix,
      },
      (err, data) => {
        // Create ../out-monoliths folder if it doesn't exist
        if (!fs.existsSync(`../out-monoliths`)) {
          fs.mkdirSync(`../out-monoliths`);
        }

        // Create ../in and ../out if they don't exist already
        if (!fs.existsSync(`../in`)) {
          fs.mkdirSync(`../in`);
        }

        if (!fs.existsSync(`../out`)) {
          fs.mkdirSync(`../out`);
        }

        // Clean the in and out folders one directory above
        if (fs.existsSync(`../in`)) {
          fs.readdirSync(`../in`).forEach((file) => {
            fs.unlinkSync(`../in/${file}`);
          });
        }
        if (fs.existsSync(`../out`)) {
          fs.readdirSync(`../out`).forEach((file) => {
            fs.unlinkSync(`../out/${file}`);
          });
        }

        const promises = [];

        // Download a single tiers demos
        filterDemoByTier(data, tierName).forEach((item) => {
          const fileName = item.Key.split("/")[item.Key.split("/").length - 1];
          const filePath = `../in/${fileName}`;
          console.log(`Downloading ${fileName}`);
          const promise = new Promise((res) => {
            s3.getObject(
              {
                Bucket: "cscdemos",
                Key: item.Key,
              },
              (err, data) => {
                // Save this to ./tierName/filename
                console.log(`Saving ${fileName} to ${filePath.slice(0, -4)}`);
                return JSZip.loadAsync(data.Body).then((zip) => {
                  zip
                    .file(fileName.slice(0, -4))
                    .async("nodebuffer")
                    .then((content) => {
                      fs.writeFileSync(filePath.slice(0, -4), content);
                      res();
                    });
                });
              }
            );
          });
          promises.push(promise);
        });

        Promise.all(promises).then(() => {
          // Run the go program from one directory above
          if (promises.length === 0) {
            console.log("No demos found for this tier");
            res();
            return;
          } else {
            console.log("Running go program...");
            setTimeout(() => {
              const { exec } = require("child_process");
              exec("go run .", { cwd: "../" }, (err, stdout, stderr) => {
                if (err) {
                  console.error(`Error with stats parser\n ${err}`);
                  console.error(stderr);
                  return;
                }

                // Run the python script to generate monolith.py
                console.log("Running python script...");
                exec(
                  "python stitch_csvs.py",
                  { cwd: "../", timeout: 1000 * 60 * 5 },
                  (err, stdout, stderr) => {
                    // Move the monolith.py to root and name it tierName.csv
                    console.log("Moving monolith.csv to root...");
                    fs.renameSync(
                      "../out/monolith.csv",
                      `../out-monoliths/${bucketPrefix.replaceAll(
                        "/",
                        "-"
                      )}-${tierName}.csv`
                    );
                    res();
                  }
                );
              });
            }, 5000);
          }
        });
      }
    );
  });
}

async function main() {
  const tiers = ["Premier", "Elite", "Challenger", "Contender", "Prospect"];
  // Process each tier
  for (const tier in tiers) {
    await processTier(tiers[tier], "s10/combines/combines-04");
  }
}

main();
