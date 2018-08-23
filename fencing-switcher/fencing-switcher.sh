#!/bin/sh

kubectl label node m1c4 --overwrite fencing/moonshot=True

kubectl label node m1c4 --overwrite fencing/moonshot=False
