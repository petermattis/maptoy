# DO NOT USE: hash table experimentation

Comparison of the builtin Go map and Robin Hood hashing.

```
name                   time/op
GoMapInsert-8           119ns ± 0%
RobinHoodInsert-8      66.0ns ± 0%
GoMapLookupHit-8       78.4ns ± 0%
RobinHoodLookupHit-8   45.0ns ± 1%
GoMapLookupMiss-8      77.6ns ± 1%
RobinHoodLookupMiss-8  44.3ns ± 0%
```

Old is `GoMap` and new is `RobinHood`.

```
name             old time/op  new time/op  delta
MapInsert-8       119ns ± 0%    66ns ± 0%  -44.55%  (p=0.000 n=8+10)
MapLookupHit-8   78.4ns ± 0%  45.0ns ± 1%  -42.59%  (p=0.000 n=9+10)
MapLookupMiss-8  77.6ns ± 1%  44.3ns ± 0%  -42.88%  (p=0.000 n=9+9)
```
