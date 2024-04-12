#include <EEPROM.h>
#include <Printers.h>
#include <XBee.h>
#include <SoftwareSerial.h>

#define LEN_SERIAL 7

#define SPEED 9600

#define TIMEOUT 300

// NOTE!! XB LIB STOPS WORKING IF VALUE TOO HIGH E.G. 60K NOT WORKING!
#define RESP_TMOUT 10000
#define RESP_TMOUT_LOW 5000
#define RESP_TMOUT_FAIL 1000

// Set TX pin number
#define XB_RX 2
// Set RX pin number
#define XB_TX 3

// Special addresses
// 0 0 => to coordinator
// 0 FFFF => to broadcast
#define ADDR_HIGH 0x0
#define ADDR_LOW 0x0

// XB TX to Ardu RX and vice versa
SoftwareSerial ss(XB_RX, XB_TX);

XBee xb = XBee();

XBeeAddress64 addr = XBeeAddress64(ADDR_HIGH, ADDR_LOW);

AtCommandRequest atrq = AtCommandRequest();
AtCommandResponse atrp = AtCommandResponse();

Tx16Request req16 = Tx16Request();
Rx16Response resp16 = Rx16Response();

Tx64Request req64 = Tx64Request();
Rx64Response resp64 = Rx64Response();

TxStatusResponse tresp = TxStatusResponse();

// serial high
uint8_t sh[] = { 'S', 'H' };

// serial low
uint8_t sl[] = { 'S', 'L' };

// DH address
uint8_t dh[] = { 'D', 'H' };

// DL address
uint8_t dl[] = { 'D', 'L' };

// MY address
uint8_t my[] = { 'M', 'Y' };

// identifier
uint8_t ni[] = { 'N', 'I' };

// association status
uint8_t ai[] = { 'A', 'I' };

char snum[LEN_SERIAL];

void readSerial() {
  for (int i = 0; i < LEN_SERIAL; i++) {
    snum[i] = EEPROM.read(i);
  }
}

void handleX16Resp() {
  if (xb.readPacket(RESP_TMOUT)) {
    // Serial.print("Got reply: ");
    // Serial.print(xb.getResponse().getApiId());
    // Serial.println();
    if (xb.getResponse().getApiId() != RX_16_RESPONSE) {
      Serial.println("Not RX16 response, skipping.");
      return;
    }
    xb.getResponse().getRx16Response(resp16);
    if (resp16.isAvailable()) {
      // Serial.println("Is OK.");
      // Serial.print("Response len: ");
      // Serial.print(resp16.getDataLength());
      // Serial.println();
      //
      // Serial.print("DATA HEX: ");
      // for (uint8_t i = 0; i < resp16.getDataLength(); i++) {
      //   Serial.print(resp16.getData(i), HEX);
      //   Serial.print(" ");
      // }
      // Serial.println();
      //
      // Serial.print("DATA STR: ");
      for (uint8_t i = 0; i < resp16.getDataLength(); i++) {
        Serial.print((char)resp16.getData(i));
      }
      Serial.println();
    } else {
      Serial.println("ERROR");
    }
  } else {
    // Serial.println("Read packet timeout.");
    if (xb.getResponse().isError()) {
      Serial.print("Error reading packet.  Error code: ");
      Serial.println(xb.getResponse().getErrorCode());
    } else {
      // Serial.println("X16: No response from radio.");
      // delay(RESP_TMOUT_FAIL);
    }
  }
}

void handleX64Resp() {
  if (xb.readPacket(RESP_TMOUT)) {
    // Serial.print("Got reply: ");
    // Serial.print(xb.getResponse().getApiId());
    // Serial.println();
    if (xb.getResponse().getApiId() != RX_64_RESPONSE) {
      Serial.println("Not RX64 response, skipping.");
      return;
    }
    xb.getResponse().getRx64Response(resp64);
    if (resp64.isAvailable()) {
      // Serial.println("Is OK.");
      // Serial.print("Response len: ");
      // Serial.print(resp64.getDataLength());
      // Serial.println();
      Serial.print("DATA HEX: ");
      for (uint8_t i = 0; i < resp64.getDataLength(); i++) {
        Serial.print(resp64.getData(i), HEX);
        Serial.print(" ");
      }
      Serial.println();
      Serial.print("DATA STR: ");
      for (uint8_t i = 0; i < resp64.getDataLength(); i++) {
        Serial.print((char)resp64.getData(i));
      }
      Serial.println();
    } else {
      Serial.println("NOT OK.");
    }
  } else {
    Serial.println("Read packet timeout.");
    if (xb.getResponse().isError()) {
      Serial.print("Error reading packet.  Error code: ");
      Serial.println(xb.getResponse().getErrorCode());
    } else {
      // Serial.println("X64: No response from radio.");
      // delay(RESP_TMOUT_FAIL);
    }
  }
}

void handleAtResp(void) {
  if (xb.readPacket(RESP_TMOUT_LOW)) {
    // Serial.println("Got reply.");
    if (xb.getResponse().getApiId() != AT_COMMAND_RESPONSE) {
      Serial.println("Not AT command response.");
      return;
    }
    xb.getResponse().getAtCommandResponse(atrp);
    if (atrp.isOk()) {
      // Serial.println("Is ok.");
      // Serial.println("Response len: ");
      // Serial.print(atrp.getValueLength());
      // Serial.println();
      Serial.print("VALUE HEX: ");
      for (uint8_t i = 0; i < atrp.getValueLength(); i++) {
        Serial.print(atrp.getValue()[i], HEX);
        Serial.print(" ");
      }
      Serial.println();
      Serial.print("VALUE STR: ");
      for (uint8_t i = 0; i < atrp.getValueLength(); i++) {
        Serial.print((char)atrp.getValue()[i]);
      }
      Serial.println();
    } else {
      Serial.println("NOT OK.");
    }
  } else {
    Serial.println("Read packet timeout.");
    if (xb.getResponse().isError()) {
      Serial.print("Error reading packet.  Error code: ");
      Serial.println(xb.getResponse().getErrorCode());
    } else {
      // Serial.println("AT: No response from radio.");
      // delay(RESP_TMOUT_FAIL);
    }
  }
}

void atCommand(uint8_t* cmd) {
  atrq.clearCommandValue();
  atrq.setCommand(cmd);
  xb.send(atrq);
  handleAtResp();
  delay(500);
}

void atCommands(void) {
  atCommand(ni);
  atCommand(my);
  atCommand(dh);
  atCommand(dl);
}

void setup() {
  // Configure XBee Digital pins
  pinMode(XB_RX, INPUT);
  pinMode(XB_TX, OUTPUT);

  delay(TIMEOUT);
  Serial.println("PIN MODES SET");

  Serial.begin(SPEED);
  while (!Serial) {}
  delay(TIMEOUT);
  Serial.println("SERIAL READY");

  ss.begin(SPEED);
  while (!ss) {}
  delay(TIMEOUT);
  Serial.println("SS READY");

  xb.begin(ss);
  delay(TIMEOUT);
  Serial.println("XBEE READY");

  readSerial();
  Serial.print("SERIAL: ");
  Serial.println(snum);
  delay(TIMEOUT);

  Serial.println("SETUP READY");
}

void loop() {
  // atCommands();
  handleX16Resp();
  // handleX64Resp();
}
