import json
import random

# 可调整：最终要抽取的总人数
TARGET_SAMPLE_SIZE = 5000

def count_all_items(user):
  total = 0
  for status in ["wish", "doing", "collect", "dropped", "on_hold"]:
    total += len(user.get(status) or [])
  return total

def main():
  with open("../data/user.json", "r", encoding="utf-8") as f:
    users = json.load(f)

  categories = {
    "normal": [],
    "medium": [],
    "heavy": [],
    "core": []
  }

  for user in users:
    total_items = count_all_items(user)
    if total_items < 200:
      categories["normal"].append(user)
    elif total_items < 500:
      categories["medium"].append(user)
    elif total_items < 1000:
      categories["heavy"].append(user)
    else:
      categories["core"].append(user)

  total_valid_users = sum(len(lst) for lst in categories.values())

  print(f"总用户数: {len(users)}")
  print(f"有效用户数: {total_valid_users}")
  for k, v in categories.items():
    print(f"{k}类用户数: {len(v)}（比例: {len(v)/total_valid_users:.2%}）")

  # 计算每类抽取数量（四舍五入）
  sampled_users = []
  for k, v in categories.items():
    portion = len(v) / total_valid_users
    n_sample = round(portion * TARGET_SAMPLE_SIZE)
    sampled = random.sample(v, min(n_sample, len(v)))
    sampled_users.extend(sampled)

  random.shuffle(sampled_users)

  # 重新分配 project_id
  for new_project_id, user in enumerate(sampled_users, start=1):
    user["project_id"] = new_project_id

  with open("../data/sampled_users_by_proportion.json", "w", encoding="utf-8") as f:
    json.dump(sampled_users, f, ensure_ascii=False, indent=2)

  print(f"\n已按比例抽取 {len(sampled_users)} 个用户，保存至 sampled_users_by_proportion.json")

if __name__ == "__main__":
  main()