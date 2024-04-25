import datetime
import json
import sys
import urllib.parse
import paho.mqtt.client as mqc
import paho.mqtt.reasoncodes as mqrc
import paho.mqtt.properties as mqp
import paho.mqtt.enums as mqe
import paho.mqtt.publish as mqpub

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
    
    cfg = None
    try:
        fn = sys.argv[1]
        with open(fn, 'r', encoding="utf-8") as f:
            cfg = json.load(f)
    except Exception as ex:
        print(f"ERROR: Could not load config from file: {fn}: {ex}")

    print("Config loaded.")
    # print(json.dumps(cfg, indent=4))
    
    c = mqc.Client(
        mqc.CallbackAPIVersion.VERSION2,
        # clean_session=True, 
        # protocol=mqc.MQTTv311
    )
    
    url = urllib.parse.urlparse(cfg['server']['url'])
    print("Url:", url)
    at: str = cfg['server']['accessToken']

    def onconn(*args, **kwargs):
        print("CONNECTED:", *args, **kwargs)
        
        ret = c.subscribe("#", qos=0)
        print("Subscribe:", ret)

        # tz=pytz.timezone("Europe/Helsinki")
        pl = {
            "ts": int(datetime.datetime.now().timestamp() * 1000), # Unix Millis
            "values": {
                "PERSON_DELTA": val
            }
        }
        print("Payload:", pl)
        
        ret = c.publish("v1/devices/me/telemetry", json.dumps(pl), qos=1, retain=False)
        print("Publish:", ret, ret.mid, ret.rc)
        ret.wait_for_publish(8)
        print("Published or timeout.")
 
        dcon = c.disconnect()
        print("MQTT Disconnected:", dcon)
        

    def onpub(*args, **kwargs):
        print("ON_PUBLISH:", *args, **kwargs)


    def onsub(*args, **kwargs):
        print("ON_SUBSCRIBE:", *args, **kwargs)


    def onfail(*args, **kwargs):
        print("CONNECT_FAILED:", *args, **kwargs)

        
    def ondisc(*args, **kwargs):
        print("DISCONNECTED:", *args, **kwargs)

    def onmsg(*args, **kwargs):
        print("ON_MESSAGE:", *args, **kwargs)


    def onlog(*args, **kwargs):
        print("ON_LOG:", *args, **kwargs)


    c.on_connect = onconn
    c.on_connect_fail = onfail
    c.on_disconnect = ondisc
    c.on_message = onmsg
    c.on_publish = onpub
    c.on_subscribe = onsub
    c.on_log = onlog
    
    # c.protocol = mqc.MQTTv311
    # c.host = url.hostname
    # c.port = 230
    # c.username = at

    c.connect_timeout = 10
    c.keepalive = 300

    # Only for callbacks
    # c.user_data_set(at)
    c.username_pw_set(at)

    con = c.connect(url.hostname, 230)
    print("MQTT Connected:", con)
    
    c.loop_forever()


if __name__ == "__main__":
    main()
