import datetime
import json
import sys
import urllib.parse
import requests

base = "https://tb.peltonet.com"
api = "api/v1"
telem = "telemetry"


def main():
    if len(sys.argv) != 3:
        print("ERROR: Invalid arguments provided.")
        print(f"Usage: {sys.argv[0]}: <CONFIG_FILE> <DELTA_COUNT>")
        return
    
    val = None
    try:
        val = int(sys.argv[2])
    except Exception as ex:
        print(f"ERROR: Could not parse input as int: {ex}")
        return
    
    cfg = None
    try:
        fn = sys.argv[1]
        with open(fn, 'r', encoding="utf-8") as f:
            cfg = json.load(f)
    except Exception as ex:
        print(f"ERROR: Could not load config from file: {fn}: {ex}")
        return

    print("Config loaded.")
    # print(json.dumps(cfg, indent=4))
    
    url = urllib.parse.urlparse(cfg['server']['url'])
    print("Url:", url)
    at: str = cfg['server']['accessToken']

    # tz=pytz.timezone("Europe/Helsinki")
    pl = {
        "ts": int(datetime.datetime.now().timestamp() * 1000), # Unix Millis
        "values": {
            "PERSON_DELTA": val
        }
    }
    print("Payload:", pl)
    
    ret = requests.post(f"{base}/{api}/{at}/{telem}", json=pl, timeout=5)
    print("HTTP Resp:", ret)
    # print("Content:", ret.text)


if __name__ == "__main__":
    main()
