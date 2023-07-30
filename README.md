# Go-OctopusAgileCalculator
A calculator of when to turn on an appliance for lowest rates using Octopus Agile. Written in Go

Currently very WIP. Currently only works hardcoded. Will be extended to take arguments and to take region.

It currently doesn't know power usage per appliance and assumes constant usage throughout its runtime. The outputted price assumes 1kW / hour consumption throughout the run.

Example output:
```
3D Print Squares: 06:00 - 15:00: 0.29p
Washer + Dryer: 10:30 - 14:30: -14.58p
```