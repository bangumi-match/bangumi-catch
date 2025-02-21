import os

def delete_empty_json_files(directory):
  for root, _, files in os.walk(directory):
    for file in files:
      if file.endswith('.json'):
        file_path = os.path.join(root, file)
        if os.path.getsize(file_path) == 18:
          os.remove(file_path)
          print(f"Deleted: {file_path}")

def delete_empty_dirs(directory):
  for root, dirs, _ in os.walk(directory, topdown=False):
    for dir in dirs:
      dir_path = os.path.join(root, dir)
      if not os.listdir(dir_path):
        os.rmdir(dir_path)
        print(f"Deleted empty directory: {dir_path}")

if __name__ == "__main__":
  delete_empty_json_files("../user_data")
  delete_empty_dirs("../user_data")