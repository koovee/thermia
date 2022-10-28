![Build](https://github.com/koovee/thermia/actions/workflows/build.yml/badge.svg)
![Test](https://github.com/koovee/thermia/actions/workflows/test.yml/badge.svg)

# Thermia controller

This project controls Thermia Diplomat Duo heat pump using Shelly 1 plus relay.

Thermia:

- short 307 and 308 pins: *EVU STOP*
- use 10 kOhm resistance between 307 and 308: *ROOM LOWERING* mode

Shelly switch is connected between 307 and 308 pins.

There are two operating modes: *simple threshold* and *dynamic threshold*.

## Threshold

Heating is controlled based on `THRESHOLD` (maximum price in *c/kWh*). Heating is *OFF* or in *ROOM LOWERING* mode 
when price is higher than threshold and ON when price is lower than threshold.

## Active hours

Heating is controlled based on `ACTIVE_HOURS` (specifies the number of hours that heating must be on during a day).
Heating is *OFF* or in *ROOM LOWERING* mode when current hour is not one of the `ACTIVE_HOURS` cheapest hours of the 
day.

## Threshold and active hours

Heating is on if hour price is lower than the `THRESHOLD` or hour is one of the cheapest hours of the day. This
makes sure that heating is on at least *n* hours a day. Number of hours is specified by `ACTIVE_HOURS` environment 
variable.

## Schedule

This is fallback mode that is normally used when *spot price* information is not available. Default hours are 00-06. 
This can be overriden with `SCHEDULE` environment variable.

```
# from 00 to 06
SCHEDULE="00,01,02,03,04,05,06"
```

In case only `SCHEDULE` environment variable is specified, fallback mode is used.

# Configuration

`THRESHOLD` maximum price (*c/kWh*). Heating is ON if *spot price* is lower and OFF if *spot price* is higher than the
threshold.

`ACTIVE_HOURS` number of hours that heating must be ON



