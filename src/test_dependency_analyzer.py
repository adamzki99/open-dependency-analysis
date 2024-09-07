import unittest
import pandas as pd
import dependency_analyzer 

from unittest.mock import patch, mock_open, MagicMock

class TestFunctions(unittest.TestCase):

    @patch("builtins.open", new_callable=mock_open, read_data='{"key": "value"}')
    def test_load_json_success(self, mock_open):
        """ Test successful loading of a JSON file """
        result = dependency_analyzer.load_json("dummy.json")
        self.assertEqual(result, {"key": "value"})

    @patch("builtins.open", new_callable=mock_open)
    def test_load_json_file_not_found(self, mock_open):
        """ Test loading JSON when file is not found """
        mock_open.side_effect = FileNotFoundError
        with self.assertRaises(FileNotFoundError):
            dependency_analyzer.load_json("nonexistent.json")

    @patch("builtins.open", new_callable=mock_open, read_data='invalid json')
    def test_load_json_invalid(self, mock_open):
        """ Test loading JSON with invalid JSON structure """
        with self.assertRaises(ValueError):
            dependency_analyzer.load_json("invalid.json")

    def test_json_to_dataframe(self):
        """ Test conversion of JSON to DataFrame """
        json_data = [
            {"PackageName": "pkg1", "References": [1, 2, 3]},
            {"PackageName": "pkg2", "References": [4, 5]}
        ]
        expected_df = pd.DataFrame({"pkg1": [[1, 2, 3]], "pkg2": [[4, 5]]})
        result_df = dependency_analyzer.json_to_dataframe(json_data)
        pd.testing.assert_frame_equal(result_df, expected_df)

    @patch("builtins.open", new_callable=mock_open)
    def test_cache_dataframe(self, mock_open):
        """ Test caching DataFrame """
        df = pd.DataFrame({"pkg1": [1], "pkg2": [2]})
        dependency_analyzer.cache_dataframe(df, "cache_file.pkl")
        mock_open.assert_called_once_with("cache_file.pkl", "wb")

    @patch("builtins.open", new_callable=mock_open)
    def test_load_cached_dataframe(self, mock_open):
        """ Test loading DataFrame from cache """
        df = pd.DataFrame({"pkg1": [1], "pkg2": [2]})
        mock_open.return_value = MagicMock()
        with patch("pickle.load", return_value=df):
            result = dependency_analyzer.load_cached_dataframe("cache_file.pkl")
            pd.testing.assert_frame_equal(result, df)

if __name__ == "__main__":
    unittest.main()
