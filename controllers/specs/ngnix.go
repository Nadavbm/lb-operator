package specs

import (
	"fmt"

	configmaps "example.com/lb/apis/configmaps/v1alpha1"
)

func createNgnixConfig(cm *configmaps.ConfigMap) map[string]string {
	m := make(map[string]string)

	for k, v := range createNgnixConf(cm) {
		m[k] = v
	}

	for k, v := range createNgnixVirtualHosts(cm) {
		m[k] = v
	}

	return m
}

func createNgnixConf(cm *configmaps.ConfigMap) map[string]string {
	m := make(map[string]string)

	fileName := ""
	data := fmt.Sprintf(`    user nginx;
    worker_processes  3;
    error_log  /var/log/nginx/error.log;
    events {
      worker_connections  10240;
    }
    http {
      log_format  main
              'remote_addr:$remote_addr\t'
              'time_local:$time_local\t'
              'method:$request_method\t'
              'uri:$request_uri\t'
              'host:$host\t'
              'status:$status\t'
              'bytes_sent:$body_bytes_sent\t'
              'referer:$http_referer\t'
              'useragent:$http_user_agent\t'
              'forwardedfor:$http_x_forwarded_for\t'
              'request_time:$request_time';
      access_log	/var/log/nginx/access.log main;
      server {
          listen       80;
          server_name  _;
          location / {
              root   html;
              index  index.html index.htm;
          }
      }
      include /etc/nginx/virtualhost/virtualhost.conf;
    }`)
	m[fileName] = data
	return m
}

func createNgnixVirtualHosts(cm *configmaps.ConfigMap) map[string]string {
	m := make(map[string]string)

	fileName := ""
	data := fmt.Sprintf(`    upstream app {
		server %s:%d;
		keepalive 1024;
	  }
	  server {
		listen 80 default_server;
		root /usr/local/app;
		access_log /var/log/nginx/app.access_log main;
		error_log /var/log/nginx/app.error_log;
		location / {
		  proxy_pass http://app/;
		  proxy_http_version 1.1;
		}
	  }`, cm.Spec.DNSDomain, cm.Spec.Ports[0])
	m[fileName] = data
	return m
}
