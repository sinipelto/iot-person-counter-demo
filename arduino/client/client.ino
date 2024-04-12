#include <Printers.h>
#include <XBee.h>
#include <SoftwareSerial.h>
#include <EEPROM.h>

#define LEN_SERIAL 7

// NOTE!!! Only enable when (re-)writing serial number to EEPROM!!!
// Normally should be disabled once set!
// Ensure set to 0!!!
#define WRITE_SERIAL 0

#if WRITE_SERIAL == 1
char dser[LEN_SERIAL] = { 'U', 'N', 'O', '4', '5', '6', 0 };
#endif

// USB/Serial Baud
#define SPEED 9600

// in ms (uint16_t)
#define SLEEP 150000

// Start timeouts
#define TIMEOUT 300

#define SHORT_DELAY 10

#define RESP_TMOUT_LOW 5000

// DOUBLE CHECK HW LAYOUT
#define XB_RX 2
#define XB_TX 3

#define PIR 8

// Special addresses
// 0 0 => to coordinator
// 0 FFFF => to broadcast

// Use 64bit addressing
// Define both DH and DL to coordinator SH and SL
#define ADDR_HIGH 0x13A200
#define ADDR_LOW 0x41779551

// Use 16bit addressing
// High = 0, Low = dest MY
// #define ADDR_HIGH 0x0
// #define ADDR_LOW 0xAAAA

// Buffer size for message string and xbee message
#define BUFS 32

// NOTE! Cant print 64bit directly with sprintf
// Figure out how to print uint64
uint32_t c;
uint32_t k;

char snum[LEN_SERIAL];

char m[BUFS];
uint8_t buf[BUFS];

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

#if WRITE_SERIAL == 1
void writeSerial() {
  for (int i = 0; i < LEN_SERIAL; i++) {
    EEPROM.write(i, dser[i]);
  }
}
#endif

void readSerial() {
  for (int i = 0; i < LEN_SERIAL; i++) {
    snum[i] = EEPROM.read(i);
  }
}

void handleStatusResp() {
  if (xb.readPacket(RESP_TMOUT_LOW)) {
    // Serial.println("Got reply!");
    if (xb.getResponse().getApiId() != TX_STATUS_RESPONSE) {
      Serial.println("Not TX Status Response.");
      return;
    }
    xb.getResponse().getTxStatusResponse(tresp);
    if (!tresp.isSuccess()) {
      Serial.println("Status NOT OK.");
      Serial.print("STATUS: ");
      Serial.print(tresp.getStatus(), HEX);
      Serial.println();
    } else {
      // Serial.println("Status OK.");
    }
  } else {
    Serial.println("Read status packet timeout.");
    if (xb.getResponse().isError()) {
      Serial.print("Error reading status packet.  Error code: ");
      Serial.println(xb.getResponse().getErrorCode());
    } else {
      Serial.println("TX_STAT: No response from radio.");
    }
  }
}

void sendX16Req(uint8_t len) {
  // Update buffer length and send
  req16.setPayloadLength(len);

  // Send TX16
  xb.send(req16);
}

void sendX64Req(uint8_t len) {
  // Update buffer length and send
  req64.setPayloadLength(len);

  // Send TX64
  xb.send(req64);
}

void sendXbee(const char *msg) {
  int len = strlen(msg);

  if (len > BUFS) {
    Serial.println("ERROR: Buffer overflow. Truncating length to buffer size.");
    len = BUFS;
  }

  // Convert char[] to uint8_t[]
  for (int i = 0; i < len; i++) {
    buf[i] = (uint8_t)(msg[i]);
  }

  // Send TX16
  // sendX16Req(len);

  // Send TX64
  sendX64Req(len);

  // Optionally report status errors
  handleStatusResp();
}

void ping() {
  sprintf(m, "P_%s_%lu", snum, k++);
  Serial.println(m);
  sendXbee(m);
}

void motion() {
  sprintf(m, "M_%s", snum);
  Serial.println(m);
  sendXbee(m);
}

void setup() {
  c = 0;
  k = 0;

  pinMode(PIR, INPUT);
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

#if WRITE_SERIAL == 1
#warning "SERIAL WRITE ENABLED"
  writeSerial();
  Serial.println("SERIAL NUMBER WRITTEN OK");
#endif
  // Read device serial from EEPROM persistent flash
  readSerial();
  Serial.print("SERIAL: ");
  Serial.println(snum);
  delay(TIMEOUT);

  // Set TX16 dest addr and payload pointer
  req16.setAddress16((uint16_t)ADDR_LOW);
  req16.setPayload(buf);

  // Set TX64 dest addr and payload pointer
  req64.setAddress64(addr);
  req64.setPayload(buf);

  delay(TIMEOUT);
  Serial.println("SETUP READY");
}

void loop() {
  // If it detects moving people
  if (digitalRead(PIR)) {
    motion();
    // Sleep until the signal resets back to zero
    // so we dont re-send a message on the same signal
    while (digitalRead(PIR)) { delay(SHORT_DELAY); }
  } else {
    // Send ping packet periodically
    // so we can determine if the Arduino is still alive
    if (++c >= SLEEP / SHORT_DELAY) {
      ping();
      c = 0;
    }
    delay(SHORT_DELAY);
  }
}
