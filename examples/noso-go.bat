:loop
@echo off

Rem #################################
Rem ## Begin of user-editable part ##
Rem #################################


REM Valid Pools:
REM   DevNoso
REM   leviable
REM   russiapool

REM You can specify multiple wallet addresses, which will be
REM cycled through round-robin after each disconnect. Useful
REM If you want to maintain a single shell script for multiple
REM miners:
REM set "WALLET=leviable leviabl2 leviable3"

set "POOL=DevNoso"
set "CPU=2"
set "WALLET=devteam_donations"

Rem #################################
Rem ##  End of user-editable part  ##
Rem #################################
setlocal enableDelayedExpansion

set "wallets="
FOR %%A IN (!WALLET!) DO set wallets=!wallets! --wallet %%A

tasklist /FI "IMAGENAME eq noso-go.exe" 2>NUL | find /I /N "noso-go.exe">NUL
if "%ERRORLEVEL%"=="0" taskkill /F /im noso-go.exe

noso-go.exe mine pool !POOL! !wallets! --cpu !CPU!
timeout 10
goto loop
