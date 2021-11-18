# Quickstart - RaspberryPi
Untar the archive:

```
$ tar -zxvf noso-go-v1.6.0-linux-arm64.tgz
noso-go
README.md
noso-go.sh
```

Edit the `noso-go.sh` file, and update the POOL, WALLET, and CPU variables. For instance, 

```
POOL="devnoso"
WALLET=""
CPU=1
```

Would become like so for me:
```
POOL="devnoso"
WALLET="leviable"
CPU=12
```

*NOTE* You should set `CPU` to the maximum *PHYSICAL* cores on your computer. `noso-go` cannot use hyperthreading/hardware-threads, so setting `CPU` higher than your *PHYSICAL* cores will likely reduce your overall hashrate.

Finally, run shell script:

```
$ bash noso-go.sh
2021/11/18 14:59:52 Writing logs to: /home/user/workspace/noso-go/bin/noso-go.log
2021/11/18 14:59:52 
# ##########################################################
#
# noso-go v1.6.0beta1 by levi.noecker@gmail.com (c)2021
# https://github.com/Noso-Project/noso-go
# Commit: ade8db0d
#
# ##########################################################
2021/11/18 14:59:52 Connecting to 109.230.238.119:8082 with password UnMaTcHeD
2021/11/18 14:59:52 Using wallet address(es): leviable
2021/11/18 14:59:52 Number of CPU cores to use: 12
2021/11/18 14:59:53 Using wallet address: leviable
2021/11/18 14:59:53 -> JOIN ng1.6.0beta1
2021/11/18 14:59:53 <- JOINOK N2RKVvyf254FFSR7BZgduCkNEbzizE2 200000000 PoolData 33380 B12E8D21249D483C9A30D886BA957539 10 7 102 0 -30 1578700 2
```

NOTE: If you get a permissions error, run this command in your terminal window, then try again:
```
chmod a+x noso-go
```
