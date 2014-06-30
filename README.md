blitzer
=======

Blitzer is a control panel for your system alerts.

When you get an alert from Nagios, Blitzer gets it too. It immediately starts
collecting diagnostics relevant to the problem. Your alert links to a web page
served by Blitzer, which shows you all the latest diagnostics that have been
collected.

Don't waste precious minutes paging through graphs and typing diagnostic commands.
Blitzer will have your context ready for you as soon as you open your laptop.

Configuration
-----

Blitzer looks for its configuration under `/etc/blitzer`. There are some example
config files under the `etc` directory to get you started.

Blitzer watches for a new _event_, determines the probes _triggered_ by that
event, and executes those _probes_. Triggers are configured in
`/etc/blitzer/triggers.d/*.yaml` and probes are defined in
`/etc/blitzer/probes.d/*.yaml`.

Example
-----

Event: "Search API is CRITICAL"

Implied probes:
  * ps_by_cpu:api.*
  * graphite_cpu:api.*
  * graphite_memory:api.*
  * graphite_diskio:db.*
  * du_rootstar:api.*
  * du_rootstar:db.*

Caveats
-----

* This won't work if Ansible asks for any passwords

Todo
-----

* Change probe def arguments to a map
* Fold more stuff into template context
* Find better way to write template results to string
