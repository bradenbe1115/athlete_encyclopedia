import json

def write_to_JSON(data: list[dict], file_name: str) -> None:
    """
        Write data to JSON at specified file location

        Args:
            data (list[dict]): valid JSON
            file_name (str): fully qualified file name 
    """
    with open(file_name, "w") as f:
        json.dump(data, f)

def read_from_JSON(file_name: str) -> list[dict]:
    with open(file_name, "r") as f:
        data = json.load(f)
    return data