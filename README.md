![Build status](https://github.com/koovee/thermia/actions/workflows/test.yml/badge.svg)


# Thermia controller

This project controls Thermia Diplomat Duo heat pump using Shelly 1 plus relay.

Thermia:

- short 307 and 308 pins: *EVU STOP*
- use 10 kOhm resistance between 307 and 308: *ROOM LOWERING* mode

Shelly switch is connected between 307 and 308 pins.

There are two operating modes: *simple threshold* and *dynamic threshold*.

## simple threshold

Heat pump is controlled based on `THRESHOLD` (maximum price in *c/kWh*). Heat pump is *OFF* or in *ROOM LOWERING* mode 
when price is higher than threshold and ON when price is lower than threshold.

## dynamic threshold

If *spot prices* are lower than `THRESHOLD`, this mode operates as [simple threshold](#simple_thershold). However, if 
*spot prices* are higher than threshold, heat pump is on during the cheapest hours of the day. `ACTIVE_HOURS` specifies
the number of hours that heat pump needs to be active every day. 

# Configuration

`THRESHOLD` maximum price (*c/kWh*). Heating is ON if *spot price* is lower and OFF if *spot price* is higher than the
threshold.

`ACTIVE_HOURS` number of hours that heating must be ON



