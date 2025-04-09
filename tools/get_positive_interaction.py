import json
import math
import random

# user_json_file = '../data/user.json'
# result_file = '../data/positive_interaction.txt'
user_json_file = '../data/sampled_users_by_proportion.json'
result_file = '../data/sampled_positive_interaction.txt'

def filter_zero_rates(items):
  """Filter out items with a rate of 0."""
  return [item for item in items if item.get('rate', 0) > 0],[item for item in items if item.get('rate', 0) == 0]


def process_positive_feedback(users):
  """Process users to extract positive feedback interaction data."""
  output = []

  for user in users:
    liked = set()

    # Add 'doing' and 'wish' subjects
    for item in user.get('doing', []) or []:
      liked.add(item['project_id'])
    for item in user.get('wish', []) or []:
      liked.add(item['project_id'])

    # Process 'collect' subjects
      filtered_collect, zero_collect = filter_zero_rates(user.get('collect', []) or [])
      if filtered_collect:
        filtered_collect.sort(key=lambda x: x['rate'], reverse=True)

        n = len(filtered_collect)
        k = math.ceil(0.7 * n)
        k = min(k, n) if k > 0 else n
        threshold = filtered_collect[k - 1]['rate']

        for item in filtered_collect:
          if item['rate'] >= threshold:
            liked.add(item['project_id'])

      # Add items from zero_collect to liked
      for item in zero_collect:
        liked.add(item['project_id'])

    # If 'liked' is empty, randomly add 3 subjects from all collections
    if not liked:
      all_items = (
          (user.get('doing') or []) +
          (user.get('wish') or []) +
          (user.get('collect') or []) +
          (user.get('dropped') or []) +
          (user.get('on_hold') or [])
      )
      random.shuffle(all_items)
      for item in all_items[:3]:
        liked.add(item['project_id'])

    # Convert to sorted project_id list
    subjects = sorted(liked)

    # Build output line
    line_parts = [str(user['project_id'])]
    line_parts.extend(map(str, subjects))
    output.append(" ".join(line_parts))

  return output


def main():
  # Read user_lite.json
  with open(user_json_file, 'r', encoding='utf-8') as file:
    users = json.load(file)

  # Process positive feedback
  output = process_positive_feedback(users)

  # Write to output file
  with open(result_file, 'w', encoding='utf-8') as file:
    file.write("\n".join(output))


if __name__ == "__main__":
  main()
