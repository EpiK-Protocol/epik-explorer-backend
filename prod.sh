#!/bin/bash

sh ops.sh build prod
sh ops.sh push prod
sh ops.sh run prod

ssh centos@3.112.3.103 '
sudo docker logs epik-explorer-backend -f
'