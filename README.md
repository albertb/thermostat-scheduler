# Thermostat scheduler

Updates a [BlueLink Smart Connect](https://bluelinksmartconnect.com/) thermostat
program to pre-heat before, and lower the heat during, periods of peak
electricity demand.

In theory, it should work with any thermostat from Braeburn with the BlueLink
Smart Connect feature, but I've only tested it with the thermostat I happen to
have: [BlueLink Model 7205](https://www.braeburnonline.com/products/thermostat/7205).

![Picture of my thermostat](thermostat.png)

## Goal

I created this to handle periods of peak demand with Hydro-Québec. They have a
special rate called [Winter Credit Option](https://www.hydroquebec.com/residential/customer-space/rates/winter-credit-option.html)
where they issue a small refund whenever you can reduce your electricity usage
during periods of peak demand.

## Usage

There's two files that are used to determine how the thermostat should be
programmed: `config.yaml` and `events.yaml`.

In `config.yaml`, you define the normal program, and configure how periods of
peak demand should be handled:

```yaml

normal_program:
  sunday:
    morning: { time: 7h,  heat: 21, cool: 24 }
    day:     { time: 9h,  heat: 20, cool: 25 }
    evening: { time: 16h, heat: 21, cool: 24 }
    night:   { time: 21h, heat: 20, cool: 25 }
  // ... all other days

peak_program:
  // How long to pre-heat for before a peak period.
  pre_heat_duration: 1h
  // How much to pre-heat vs the normal temperature.
  pre_heat_temp_offset: 1
  // How much to lower the temperature during a peak period.
  peak_temp_offset: -2
```

And in `events.yaml`, you add the dates and times of periods of peak demand:

```yaml
// January 10st 2024, from 6am to 9am, and 4pm to 8pm.
- { date: 2024-01-10,
    events: [{ start: 6h , end: 9h  },
             { start: 16h, end: 20h }] }
```

Finally, run the binary before and after the peak period to update the
thermostat program. Or better yet, run it every few hours in a cronjob, and
just append to `events.yaml` whenever a new peak period is annouced.
