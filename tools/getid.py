import re

# 输入日志文件
log_file = "log_subject_20250220_114411.txt"
# 输出 ID 文件
output_file = "ids.txt"

# 正则匹配 ID
pattern = re.compile(r"Error fetching ID (\d+)")

# 读取日志并提取 ID
ids = []
with open(log_file, "r", encoding="utf-8") as f:
    for line in f:
        match = pattern.search(line)
        if match:
            ids.append(match.group(1))

# 写入 ID 到文件
with open(output_file, "w", encoding="utf-8") as f:
    f.write("\n".join(ids))

print(f"提取完成，共找到 {len(ids)} 个 ID，结果保存在 {output_file}")
