#!/bin/bash
maxblocks=3
for i in `seq $maxblocks`; do
  maxtxinblock=200
  (
  for j in `seq $maxtxinblock`; do
    vertcoin-cli sendtoaddress WqynWw2n93rfihLNsUVRKDxEBQc19rKGof 0.00001
  done
  vertcoin-cli generate 6
  ) &

  (
  for j in `seq $maxtxinblock`; do
    bitcoin-cli sendtoaddress n3k3QXbArVijHeU2HpJQSRpW8jAaD1v4Tf 0.00001
  done
  bitcoin-cli generate 6
  ) & 

  (
  for j in `seq $maxtxinblock`; do
    litecoin-cli sendtoaddress mxWrYVpUSsGEyjJaSfTjVRQYDTT3ksQ7oB 0.00001
  done 
  litecoin-cli generate 6
  )
  # You can run a whole bunch of these to test deposits -- it's a really good way to flood transactions.
  # also run lots of generate scripts alongside, and lots of nodes.

  # BTW all these addresses are just for a particular user at the time of writing -- make it adaptable to more
done