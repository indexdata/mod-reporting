# VPN cheat sheet

I use AWS VPN Client to allow mod-reporting, running locally, to access MetaDB.


## Testing the connection

If this is working correctly, it can be easily verified using `curl`:
```
$ time curl id-test-metadb.folio.indexdata.com:5432
curl: (52) Empty reply from server

real	0m0.300s
user	0m0.006s
sys	0m0.009s
$ 
```
But if it is not working, there is a seventy-five second wait before a timeout:
```
$ time curl id-test-metadb.folio.indexdata.com:5432

curl: (28) Failed to connect to id-test-metadb.folio.indexdata.com port 5432 after 75036 ms: Couldn't connect to server

real	1m15.050s
user	0m0.007s
sys	0m0.014s
$ 
```


## Troubleshooting

If the VPN is running but `curl` can't connect, it's probably a DNS problem. Verify using `host`:
```
$ host id-test-metadb.folio.indexdata.com
id-test-metadb.folio.indexdata.com is an alias for metadb-indexdata-test.cpvupbknx9nm.us-east-1.rds.amazonaws.com.
metadb-indexdata-test.cpvupbknx9nm.us-east-1.rds.amazonaws.com is an alias for ec2-34-232-193-131.compute-1.amazonaws.com.
ec2-34-232-193-131.compute-1.amazonaws.com has address 34.232.193.131
$ 
```

If the address is in the 34.*.*.*, the VPN will not be not in effect. You want to see something in the 172.* namespace.


## DNS problems

You can check your DNS sources on MacOS using System Preferencess -> Network -> Advanced -> DNS. The list of DNS servers in the left column should all be greyed out. If you have added a custom nameserver, it has to be removed, and the changes applied.

Once this is done, DNS caching will continue to defeat you, so:
* Close the VPN
* Turn off WiFi
* Clear the DNS cache with `sudo dscacheutil -flushcache; sudo killall -HUP mDNSResponder`
* Turn on the WiFi
* Open the VPN

Finally, all should be well.


