# iot-person-counter-demo

A demo set of applications for various devices to build up a people counter tracking system

to track the number of people in a room per time.

## The software components included in the repo:
* Arduino sensor + xbee client sensor tracker application (Arduino C++)
* Arduino xbee server receiver application (Arduino C++)
* Raspberry Pi edge serial-to-cloud server application (GoLang)

## Following external software components are required:
* OS for Raspberry Pi: e.g. Ubuntu Server 22.04 LTS or similar Linux
* Arduino IDE for compiling and uploading Arduino sketches
* Simple Notepad/text editor for editing the configuration files
* ThingsBoard IoT Platform server instance (Cloud instance or Self-hosted) => Docker and docker-compose recommended for self-hosting
* Digi XCTU tool for flashing Xbee firmware & configuration
* Virtual Serial Port Emulator for emulating the serial port communication on local development setup
* GoLang toolchain and compiler for target platform (e.g. Linux ARM64) to compile the edge application for Raspi
*(VSCode or similar IDE for locally developing the Raspi edge application)

## Minimal setup (1 doorway) consists of following hardware:
* 3x Grove Sensor Boards
* 2x Grove PIR sensors
* 3x Digi Xbee V2 Boards
* 3x Digi Xbee S1 modem modules
* 1x Digi Xbee Explorer device or similar FTDI USB chip for Xbee modem firmware flash & configuration
* 3x Arduino MCU chips (UNO Rev3 or similar)
	* 2x clients with 1 PIR sensor and 1 Xbee client/transmitter module connected to each
	* 1x server with Xbee server/coordinator module connected
* 1x Raspberry Pi 3B+ / 4 / 5
	* 1x 1 SD Card 4GB+
	* 1x Micro-USB with power adapter 5V 2A+
* 3x USB-A=>USB-B cables (Arduino power)
* 1x mini-USB cable (Xbee Explorer)

1. First connect the Xbee boards to the Arduinos
2. Connect 2 Grove boards to the 2 client Arduinos
3. Connect Xbee modem to the Xplorer chip and that to PC
4. Flash the correct Xbee firmware, and apply the profile files for the 2 clients and 1 server/coordinator, respectively
5. Connect the Xbee modules to the Xbee boards
6. Using PC, flash the client application to 2 client arduinos
7. For the first time, enable the "serial write" option in the arduino sketch, and flash all the arduinos with unique identifiers each to the EEPROM
8. After that, disable serial write and re-flash firmware
9. Flash the server application to the 3rd server arduino, same serial stuff, make sure each arduino has unique serialnum
10. Place the two sensor arduinos on both sides of a doorway if no door exists or both on the side where door does not open
11. Connect the arduinos to a power source
12. Connect the third server arduino to the raspberry pi USB port
13. Set up Raspberry Pi with proper OS e.g. Raspberry Pi OS or Ubuntu Server
14. Insert the flashed sd card to raspi, Set up Raspi with SSH access
15. Create a configuration file for the application as `config.linux.json` with proper details. E.g. serial port, device serial and to determine correct sensors locations
16. In the deploy.ps1 script, set proper hostname and configure SSH config with ssh Host name, keys etc.
17. Build the Go application for Linux arm64 and upload the binary and configuration to the Raspi using the deploy.ps1 script (Windows/Powershell)
18. Start the application e.g. as a systemd service
19. Make sure your ThingsBoard cloud environment is set up and starts receiving telemetry data

