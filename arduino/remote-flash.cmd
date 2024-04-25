set bd=build\arduino.avr.uno
set srv1=raspitf
set srv2=raspit

if [%1] == [server] (
goto :server
)

:client
set tg=client
pushd %tg%\%bd%
scp %tg%.ino.hex %srv1%:~/%tg%.hex
ssh %srv1% -tC "bin/arduino-cli upload -i %tg%.hex -p /dev/ttyACM0"
ssh %srv1% -tC "bin/arduino-cli upload -i %tg%.hex -p /dev/ttyACM1"
popd

if [%1] == [client] (
goto :end
)

:server
set tg=server
pushd %tg%\%bd%
scp %tg%.ino.hex %srv2%:~/%tg%.hex
ssh %srv2% -tC "bin/arduino-cli upload -i %tg%.hex -p /dev/ttyACM0"
popd

:end
timeout 5
exit
