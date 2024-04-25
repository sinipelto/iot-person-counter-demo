import random
import threading
from serial import *
import serial
import time

port = 'COM1'
baud = 9600
bs = 8
parity = 'N'
stop = 1
tmout = 5.0

rng = random.SystemRandom()
ser = serial.Serial(port, baud, bs, parity, stop, tmout, True)

def msg(data):
    print(data)
    ser.write(data.encode())

t1 = lambda i: threading.Thread(target=msg, args=(f"P_WIN3201_{i}\r\n",))
t2 = lambda i: threading.Thread(target=msg, args=(f"P_WIN3202_{i}\r\n",))

msgs = ["M_WIN3201\r\n", "M_WIN3202\r\n"]

ths: list[threading.Thread] = [None, None]

def randFloat(min: float, max: float):
    if min > max:
        raise ValueError("MIN > MAX")
    return rng.random() * (max - min) + min

def ping(x):
    ths = [t1(x), t2(x)]
    rng.shuffle(ths)
    for t in ths:
        t.start()
    time.sleep(randFloat(8,20))
    for t in ths:
        t.join(3)

print("START")

try:
        c = 0
        while not ser.is_open and c <= 50:
            c+=1
            try:
                ser.open()
            except Exception as ex:
                print("Could not open serial:", ex)
                time.sleep(3)
        print("Serial open.")

        ser.flush()
        time.sleep(0.5)
        ser.reset_input_buffer()
        time.sleep(0.5)
        ser.reset_output_buffer()
        time.sleep(0.5)
        print("Serial flushed.")

        print()

        i = 0
        j = 0
        
        for x in range(500):
            for j in range(i, i+2):
                ping(j)
                i = j

            for y in range(6):
                rng.shuffle(msgs)
                print("0:")
                msg(msgs[0])
                time.sleep(randFloat(0.1, 5))
                print("1:")
                msg(msgs[1])
                time.sleep(randFloat(3,8))

            for y in range(i, i+2):
                ping(j)
                i = j

except KeyboardInterrupt:
    print("Interrupted.")
    for t in ths:
        if t:
         t.join(3)
    print("Joined.")
    ser.close()
    print("Closed.")

print("EXIT")
