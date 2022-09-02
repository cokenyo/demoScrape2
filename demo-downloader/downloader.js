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
  return data.Contents.filter((item) => item.Key.includes(tierName));
}

s3.listObjectsV2(
  {
    Bucket: "cscdemos",
    Prefix: "s9/Combines/Combines-02",
  },
  (err, data) => {
    const tierName = "Elite";
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

    // Download a single tiers demos
    filterDemoByTier(data, tierName).forEach((item) => {
      const fileName = item.Key.split("/")[item.Key.split("/").length - 1];
      const filePath = `../in/${fileName}`;
      console.log(`Downloading ${fileName}`);
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
              });
          });
        }
      );
    });
  }
);
