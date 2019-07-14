# 08 Helm

At this point I'm probably beating a dead horse. I'm trying to poke at all of the things I can think of that might use ZTS. I could make it better or I could make a Helm chart.

## The Chart

I cleaned up the manifests and made things a bit more programmatic. I thought about splitting it into subcharts but that seemed like a lot of work. Sprig doesn't offer much in the way of scripting on top of templating (AFAIK - if I'm wrong I'd love to know) so it's frustrating to mess around with. I guess I'm spoiled by Jinja.

