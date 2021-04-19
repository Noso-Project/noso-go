:loop

Rem #################################
Rem ## Begin of user-editable part ##
Rem #################################

REM Setting CPU to 0 will make it automatically use MAXCORES / 2
REM On my 4 core, 8 threaded i7 that means it will use all 4 physical cores
REM If you set this to something other than 0, it should be to however many
REM physical cores you have, not logical
set "CPU=0"
set "WALLET=N2RUEEpVEyF9fgmQg6HKcrwkm4MERDx"
set "ADDRESS=noso.dukedog.io"
set "PORT=8082"
set "PASSWORD=duke"

Rem #################################
Rem ##  End of user-editable part  ##
Rem #################################
@echo off
setlocal enableDelayedExpansion

taskkill /F /im noso-go.exe
timeout 3

noso-go.exe mine --address !ADDRESS! --port !PORT! --password !PASSWORD! -wallet !WALLET! -cpu !CPU!
timeout 10
goto loop
