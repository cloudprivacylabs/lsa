---
title: Measurements
---

# Measurements

Measurements are special value types that are composed of a value and
a unit. LSA provides methods to ingest and process measurements, and
allows hooks for third-party packages to translate measured valued
between different units.

Measured values can be ingested in one of the following ways:

  * A node whose value is a string containing both a measured value
    and unit. Example: 12 kg.
  * A node containing value and unit in its properties as two separate
    fields.
  * A node containing the measured value, and another node containing
    the unit.

LSA normalizes a measure value by creating a new node based on a
schema node that contains the following properties:

  * Node label includes `https://lschema.org/Measure`
  * `https://lschema.org/valueType: https://lschema.org/Measure`
  * `https://lschema.org/measure/value` contains the measured value
  * `https://lschema.org/measure/unit` contains the unit
  * `https://lschema.org/nodeValue` contains the concatenation of
    measured value and unit with a space in between.

