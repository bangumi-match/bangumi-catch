import json
import random
from collections import defaultdict

# 输入和输出文件路径
input_file = '../data/sampled_positive_interaction.txt'
train_file = '../data/train_interaction.txt'
test_file = '../data/test_interaction.txt'

def split_interactions(input_file, train_file, test_file):
  # 读取交互数据
  with open(input_file, 'r', encoding='utf-8') as f:
    lines = f.readlines()

  # 按用户分组交互数据
  user_interactions = defaultdict(list)
  for line in lines:
    parts = line.strip().split()
    user_id = parts[0]
    interactions = parts[1:]
    user_interactions[user_id].extend(interactions)

  # 初始化训练集和测试集
  train_set = defaultdict(list)
  test_set = defaultdict(list)

  # 分配交互数据
  for user_id, interactions in user_interactions.items():
    # 将用户的交互数据随机打乱以确保随机分配
    shuffled_interactions = interactions.copy()
    random.shuffle(shuffled_interactions)

    # 分配数据到训练集和测试集
    train_size = int(0.8 * len(shuffled_interactions))
    train_set[user_id] = shuffled_interactions[:train_size]
    test_set[user_id] = shuffled_interactions[train_size:]

    # Ensure training set has at least 3 unique interactions
    unique_train = list(set(train_set[user_id]))
    if len(unique_train) < 3:
      train_set[user_id] = list(set(interactions))

    # Ensure testing set has at least 3 unique interactions
    unique_test = list(set(test_set[user_id]))
    if len(unique_test) < 3:
      test_set[user_id] = list(set(interactions))

  # 写入训练集文件
  with open(train_file, 'w', encoding='utf-8') as f:
    for user_id, interactions in train_set.items():
      f.write(f"{user_id} {' '.join(interactions)}\n")

  # 写入测试集文件
  with open(test_file, 'w', encoding='utf-8') as f:
    for user_id, interactions in test_set.items():
      f.write(f"{user_id} {' '.join(interactions)}\n")

  print(f"训练集和测试集已生成：\n训练集: {train_file}\n测试集: {test_file}")

if __name__ == "__main__":
  split_interactions(input_file, train_file, test_file)