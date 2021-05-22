@echo off

Rem #################################
Rem ## Begin of user-editable part ##
Rem #################################


REM Menu code initially provided by Grumpy
REM on the Discord server

REM Valid Pools:
REM   DevNoso
REM   dukedog.io
REM   mining.moe
REM   russiapool

:menu

echo Valid Pools:
echo   1 DevNoso
echo   2 dukedog.io
echo   3 mining.moe
echo   4 russiapool

set /p id=Enter ID:

IF "%id%"=="1" (
  set "POOL=DevNoso"
  goto menudone
)
IF "%id%"=="2" (
  set "POOL=dukedog.io"
  goto menudone
)
IF "%id%"=="3" (
  set "POOL=mining.moe"
  goto menudone
)
IF "%id%"=="4" (
  set "POOL=russiapool"
  goto menudone
)

echo You have to choose between 1 and 4

goto menu

:menudone

echo You have chosen to mine at %POOL%
REM set "POOL=yzpool"
set "CPU=2"
echo giving the miner the use of %CPU% threads
set "WALLET=N2RUEEpVEyF9fgmQg6HKcrwkm4MERDx"
echo and gains to be deposit in %WALLET%

Rem #################################
Rem ##  End of user-editable part  ##
Rem #################################
:loop
setlocal enableDelayedExpansion

tasklist /FI "IMAGENAME eq noso-go.exe" 2>NUL | find /I /N "noso-go.exe">NUL
if "%ERRORLEVEL%"=="0" taskkill /F /im noso-go.exe

noso-go.exe mine pool !POOL! --wallet !WALLET! --cpu !CPU!
timeout 10
goto loop
