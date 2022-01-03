import pandas as pd
from glob import glob
import os

output_name = 'monolith.csv'
files = glob(os.path.join("out", "*.csv"))
df = None
for file in files:
    if output_name in file:
        continue
    print('adding',file)
    df_tmp = pd.read_csv(file)
    df = df_tmp if df is None else pd.concat([df,df_tmp])

df.to_csv(os.path.join("out", output_name),index=False)
