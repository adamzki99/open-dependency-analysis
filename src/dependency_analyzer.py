import json
import os
import pickle
import dependency_visualizer
import argparse
import pandas as pd

def load_json(file_path):
    """
    Load data from a JSON file.

    Args:
        file_path (str): Path to the JSON file.

    Returns:
        dict: Parsed data from the JSON file.
    """
    try:
        with open(file_path, 'r') as file:
            return json.load(file)
    except FileNotFoundError:
        raise FileNotFoundError(f"Error: The file '{file_path}' was not found.")
    except json.JSONDecodeError:
        raise ValueError(f"Error: The file '{file_path}' contains invalid JSON.")
    except Exception as e:
        raise Exception(f"An unexpected error occurred: {e}")


def json_to_dataframe(json_data: list) -> pd.DataFrame:
    """
    Convert a list of dictionaries into a pandas DataFrame.

    Args:
        json_data (list): List of dictionaries where each item represents data.

    Returns:
        pd.DataFrame: A pandas DataFrame created from the list.
    """
    data_dict = {item['Package']: item['CalledBy'] for item in json_data}
    return pd.DataFrame.from_dict(data_dict)


def cache_dataframe(df: pd.DataFrame, cache_path: str):
    """
    Cache a pandas DataFrame to a file using pickle.

    Args:
        df (pd.DataFrame): The DataFrame to cache.
        cache_path (str): Path to the cache file.
    """
    with open(cache_path, 'wb') as f:
        pickle.dump(df, f)


def load_cached_dataframe(cache_path: str) -> pd.DataFrame:
    """
    Load a cached pandas DataFrame from a pickle file.

    Args:
        cache_path (str): Path to the cache file.

    Returns:
        pd.DataFrame: Loaded DataFrame from cache.
    """
    with open(cache_path, 'rb') as f:
        return pickle.load(f)


def main(args):
    json_file_path = 'data.json'
    cache_df_file_path = '.da_df_cache'
    csv_file_name = 'output_data.csv'

    json_data = load_json(json_file_path)


    if os.path.exists(cache_df_file_path) and not args.no_cache:
        df = load_cached_dataframe(cache_df_file_path)
    else:
        print(f'Unable to find cache file "{cache_df_file_path}", reading "{json_file_path}" to build cache')
        
        df = json_to_dataframe(json_data)
        cache_dataframe(df, cache_df_file_path)

        print("Cache updated.")

    if args.create_csv:
        df.to_csv(csv_file_name, index=False)
        print(f'Data saved to "{csv_file_name}"')

    if args.create_network_view:
        dependency_visualizer.create_real_network_view(json_data, int(args.limit), args.dynamic_node_size, args.dynamic_edge_size)

if __name__ == "__main__":

    parser = argparse.ArgumentParser()

    parser.add_argument("--no_cache", action='store_true',
                        help="forces the usage of no-cached values")
    
    parser.add_argument("--create_csv", action='store_true',
                        help="create a CSV file to use as a dependency matrix")

    parser.add_argument("--create_network_view", action='store_true', 
                        help="include to create an interactive network view of the dependencies")
    parser.add_argument("--dynamic_node_size", action='store_true', 
                        help='include to enable the usage of dynamic node sizes in the network view')
    parser.add_argument("--dynamic_edge_size", action='store_true',
                        help='include to enable the usage of dynamic edge sizes in the network view')
    parser.add_argument("-l", "--limit", 
                        help="set a limit for the count of references to a given package/module to be included")
                

    main(parser.parse_args())