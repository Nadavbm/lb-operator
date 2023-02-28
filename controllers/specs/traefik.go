package specs

import (
	"fmt"

	configmaps "example.com/lb/apis/configmaps/v1alpha1"
)

func createTraefikConfig(cm *configmaps.ConfigMap) map[string]string {
	m := make(map[string]string)

	for k, v := range createTraefikStaticConfiguration(cm) {
		m[k] = v
	}

	for k, v := range createTraefikDynamicConfiugration(cm) {
		m[k] = v
	}

	return m
}

func createTraefikStaticConfiguration(cm *configmaps.ConfigMap) map[string]string {
	m := make(map[string]string)

	fileName := "static.yml"
	data := fmt.Sprintf(`# Traefik static configuration file (/config/traefik.yml)
	log:
      level: INFO
    accessLog: {}
    providers:
      file:
        filename: /config/dynamic.toml
        watch: true
      kubernetesIngress:
        allowEmptyServices: true
        namespaces: []
    entryPoints:
      web:
        address: ':80'
        http:
          redirections:
            entryPoint:
              to: websecure
              scheme: https
        transport:
          lifeCycle:
            requestAcceptGraceTimeout: 42
            graceTimeOut: 42
      websecure:
        address: ':443'
        http:
          tls: {}
        transport:
          lifeCycle:
            requestAcceptGraceTimeout: 42
            graceTimeOut: 42
      ping:
        address: ':8082'
    ping:
      entryPoint: ping
      terminatingStatusCode: 204 `)

	m[fileName] = data
	return m
}

func createTraefikDynamicConfiugration(cm *configmaps.ConfigMap) map[string]string {
	m := make(map[string]string)

	fileName := "dynamic.yml"
	data := fmt.Sprintf(`# Traefik dynamic configuration file
	  http:
		middlewares:
		  compress:
			compress: {}
		routers:
		  web:
			entryPoints:
			  - web
			middlewares:
			  - compress@file
			service: lb
			rule: '%s'
		  websecure:
			entryPoints:
			  - websecure
			middlewares:
			  - compress@file
			service: lb
			rule: '%s'
			tls:
			  options: myTLSOptions
		services:
		  lb:
			loadBalancer:
			  sticky:
				cookie:
				  name: my_sticky_cookie_name
				  secure: true
				  httpOnly: true
				  sameSite: none
	  services: {}
	  tls:
		certificates:
		  - certFile: /data/cert/tls.crt
			keyFile: /data/cert/tls.key
		stores:
		  default:
			defaultCertificate:
			  certFile: /data/cert/tls.crt
			  keyFile: /data/cert/tls.key
		options:
		  myTLSOptions:
			minVersion: VersionTLS12
			cipherSuites:
			  - TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
			  - TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384
			  - TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305
			  - TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA
			  - TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA
			  - TLS_RSA_WITH_AES_128_GCM_SHA256
			  - TLS_RSA_WITH_AES_256_GCM_SHA384
			  - TLS_RSA_WITH_AES_128_CBC_SHA
			  - TLS_RSA_WITH_AES_256_CBC_SHA
			  - TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256
			  - TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256
			  - TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256
	`, cm.Spec.DNSDomain, cm.Spec.DNSDomain)

	m[fileName] = data
	return m
}
