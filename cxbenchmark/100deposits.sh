#!/bin/bash
maxblocks=20
for i in `seq $maxblocks`; do
  maxtxinblock=100
  (
  for j in `seq $maxtxinblock`; do
    vertcoin-cli sendtoaddress WqynWw2n93rfihLNsUVRKDxEBQc19rKGof 0.00001 >/dev/null
  done
  ) &
  ch1=$!
  # Because bitcoin is really slow at this for some reason
  bitcoinblocks=$(expr $maxtxinblock / 1)
  (
  for j in `seq $bitcoinblocks`; do
    bitcoin-cli sendtoaddress n3k3QXbArVijHeU2HpJQSRpW8jAaD1v4Tf 0.00001 >/dev/null
  done
  ) &
  ch2=$!
  (
  for j in `seq $maxtxinblock`; do
    litecoin-cli sendtoaddress mxWrYVpUSsGEyjJaSfTjVRQYDTT3ksQ7oB 0.00001 >/dev/null
  done
  ) & 
  ch3=$!
  # You can run a whole bunch of these to test deposits -- it's a really good way to flood transactions.
  # also run lots of generate scripts alongside, and lots of nodes.

  # BTW all these addresses are just for a particular user at the time of writing -- make it adaptable to more
  wait $ch1
  wait $ch2
  wait $ch3
  bitcoin-cli generate 1 &
  ch1=$!
  litecoin-cli generate 1 &
  ch2=$!
  vertcoin-cli generate 1 &
  ch3=$!

  wait $ch1
  wait $ch2
  wait $ch3
  echo "Done with $maxtxinblock deposits per block for $i blocks"
done

wait $ch1
wait $ch2
wait $ch3
echo "Done with stress test!"
