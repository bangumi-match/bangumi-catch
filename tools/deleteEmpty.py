import os

def delete_empty_json_files(directory):
  for root, _, files in os.walk(directory):
    for file in files:
      if file.endswith('.json'):
        file_path = os.path.join(root, file)
        if os.path.getsize(file_path) == 18:
          os.remove(file_path)
          print(f"Deleted: {file_path}")

if __name__ == "__main__":
  delete_empty_json_files("../user_data")