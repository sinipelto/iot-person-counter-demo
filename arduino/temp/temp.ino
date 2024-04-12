int buf[64];
int *p;

void printPorts(void) {
  Serial.println("***************");
  for (uint8_t i = 0; i < 20; i++) {
    Serial.print("Pin ");
    Serial.print(i);
    Serial.print(" Analog: ");
    Serial.print(analogRead(i));
    Serial.print(" Digital: ");
    Serial.print(digitalRead(i));
    Serial.print("\n");
  }
  Serial.println("***************");
}

int *recvStream(Stream &ser) {
  cnt = 0;
  while (!ser.available()) {
    if (cnt++ > TIMEOUT) {
      Serial.println("RECV_TIMEOUT");
      return 0;
    }
    delay(10);
  }

  p = buf;
  while (ser.available()) {
    *(p++) = ser.read();
    delay(10);
  }
  *(++p) = 0;

  return buf;
}

void readSerial(int n) {
  cnt = 0;
  while (!ss.available()) {
    if (cnt++ > TIMEOUT) {
      Serial.println("READ_TIMEOUT");
      return;
    }
    delay(10);
  }
  Serial.print("RESP");
  Serial.print(n);
  Serial.println();
  while (ss.available()) {
    Serial.print(ss.read(), HEX);
    Serial.print(" ");
    delay(10);
  }
  Serial.println();
}

void setup() {
  // put your setup code here, to run once:

}

void loop() {
  // put your main code here, to run repeatedly:

}
