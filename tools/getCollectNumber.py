import os

def count_collect_files(directory):
  total_files = 0
  more_than_100 = 0
  more_than_1000 = 0
  more_than_10000 = 0

  for root, _, files in os.walk(directory):
    for file in files:
      if file == '2_collect.json':
        total_files += 1
        file_path = os.path.join(root, file)
        with open(file_path, 'r', encoding='utf-8') as f:
          lines = sum(1 for _ in f)
          if lines > 100:
            more_than_100 += 1
          if lines > 1000:
            more_than_1000 += 1
          if lines > 10000:
            more_than_10000 += 1

  return total_files, more_than_100, more_than_1000, more_than_10000

if __name__ == "__main__":
  directory = "../user_data"
  total_files, more_than_100, more_than_1000, more_than_10000 = count_collect_files(directory)
  print(f"Total '2_collect.json' files: {total_files}")
  print(f"Files with more than 100 lines: {more_than_100}")
  print(f"Files with more than 1000 lines: {more_than_1000}")
  print(f"Files with more than 10000 lines: {more_than_10000}")