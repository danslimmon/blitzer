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
=====

Blitzer looks for its configuration under `/etc/blitzer`. There are some example
config files under the `etc` directory to get you started.

Blitzer watches for a new _event_, determines the probes _triggered_ by that
event, and executes those _probes_. Triggers are configured in
`/etc/blitzer/triggers.d/*.yaml` and probes are defined in
`/etc/blitzer/probes.d/*.yaml`.

`blitzer.yaml` config file has the following global parameters:

* `addr`: The IP address on which the web UI should be served. `0.0.0.0` for
  all IP addresses.
* `port`: The port on which the web UI should be served.
* `debug`: Whether to print verbose output.

There are also probe-specific parameters defined in `blitzer.yaml`:

* `graphite` => `base_url`: The base URL for Graphite image rendering, e.g.
  `http://graphite.example.com/render`
* `ansible` => `inventory`: The [Ansible](http://ansible.com) inventory file
  to be used by the Ansible Probe.

Triggers
-----

Any event that arrives from your monitoring system can match one or more
__triggers__. A trigger says "When you see an alert that looks like this, start
running these probes and reporting their results."

Here's an example trigger configuration:

```yaml
---
service_match: '^Search API$'
probes:
  - name: 'ansible_ps_by_cpu'
    args:
      HostPattern: '^api-01$'
      LineCount: 6
  - name: 'graphite_collectd_cpu'
    args:
      Hostname: 'api-01'
```

With this trigger, Blitzer will look for alerts on the service "Search API"
(`service_match` is a regular expression). Whenever that service is down,
Blitzer will repeatedly run the `ansible_ps_by_cpu` probe and the
`graphite_collectd_cpu` probe.

The arguments passed to each probe will be available to the running probe.

Probes
-----

A __probe__ is a job that runs regularly and reports its output to Blitzer.
Blitzer supports several different sorts of probes. Each probe type takes
the following generic configuration parameters in addition to any
probe-specific parameters described below:

* `type`: The type of the probe; see subsections for a list of probe types.
* `name`: A string used to identify the probe in a trigger.
* `title`: A nicer-looking string to use when displaying probe output to
  the user.
* `interval`: How often to run the probe, in seconds
* `html`: An HTML template (of the [Go variety](http://golang.org/pkg/html/template/))
  that can be used to format the probe's output in the web UI.


h3. Graphite Probe

The Graphite Probe displays a Graphite graph. See `etc/graphite_collectd_cpu.yaml`
for an example. This probe takes the following argument:

* `qs_template`: A template that will be executed to form the Graphite query-string

In addition, `graphite_baseurl` must be set in `blitzer.yaml` to use this probe.


h3. Ansible Probe

The Ansible Probe runs an [Ansible](http://ansible.com) playbook and prints the
output of any `shell` commands in it. You can find an example of this probe in
`etc/ansible_ps_by_cpu.yaml`. It takes the following argument:

* `tasks`: A YAML list defining the playbook tasks to run.


Caveats
=====

* Won't work if Ansible asks for any passwords

Todo
=====

* Find better way to write template results to string (strings.Writer?)
* Change `debug` to a boolean
* Use httptest
* Is "middleware" what I want?
* Add standard way of testing that an HTTP request body produces a given code
