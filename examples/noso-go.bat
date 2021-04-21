:loop
@echo off

Rem #################################
Rem ## Begin of user-editable part ##
Rem #################################

REM Setting CPU to 0 will make it automatically use MAXCORES / 2
REM On my 4 core, 8 threaded i7 that means it will use all 4 physical cores
REM If you set this to something other than 0, it should be to however many
REM physical cores you have, not logical
set "POOL=yzpool"
set "CPU=2"
set "WALLET=N2RUEEpVEyF9fgmQg6HKcrwkm4MERDx"

Rem #################################
Rem ##  End of user-editable part  ##
Rem #################################
setlocal enableDelayedExpansion

tasklist /FI "IMAGENAME eq noso-go.exe" 2>NUL | find /I /N "noso-go.exe">NUL
if "%ERRORLEVEL%"=="0" taskkill /F /im noso-go.exe

noso-go.exe mine pool !POOL! --wallet !WALLET! --cpu !CPU!
timeout 10
goto loop
