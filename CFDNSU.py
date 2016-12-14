#!/usr/bin/env python

#/usr/bin/CFDNSU.py
#/etc/CFDNSU.conf
#/lib/systemd/system/CFDNSU-daemon.service
#/var/log/CFDNSU.log

import ConfigParser, requests, json, sys, time, argparse, time, datetime, os.path

parser = argparse.ArgumentParser(description='CloudFlare DNS Updater')
parser.add_argument('-c', dest='config', default='/etc/CFDNSU.conf', metavar='file', help='Use an alternative configuration file.')
parser.add_argument('-l', dest='log', default=False, metavar='file', help='Log file')
parser.add_argument('--dump', dest='dump_mode', action='store_true', required=False, help='Dumping mode')
parser.add_argument('-u', dest='email', default=False, required=False, help='Email')
parser.add_argument('-p', dest='api_key', default=False, required=False, help='API key')

args = parser.parse_args()

config = ConfigParser.RawConfigParser()

u = args.email
k = args.api_key

if os.path.isfile(args.config) and config.read(args.config) and config.has_section('configuration'):
	if args.email is False:
		u = config.get('configuration', 'x-auth-email')

	if args.api_key is False:
		k = config.get('configuration', 'x-auth-key')

	z = config.get('configuration', 'zone_identifier')
	i = config.get('configuration', 'identifier')
	s = config.get('configuration', 'subdomain')
	d = config.get('configuration', 'domain')
	l = config.get('configuration', 'log_file')
	r = config.getint('configuration', 'refresh_rate')

h = { 'X-Auth-Email' : u, 'X-Auth-Key' : k, 'Content-Type' : 'application/json' }

logWriter = open(l if not args.log else args.log, 'a')

def logWrite(text):
	text = datetime.datetime.fromtimestamp(time.time()).strftime('%Y-%m-%d %H:%M:%S - ') + text
	logWriter.write(text + "\n")
	print(text)

class CloudFlare:
	def getCloudFlareIp(self, domain):
		r = requests.get('https://api.cloudflare.com/client/v4/zones/' + z + '/dns_records', headers=self.headers)

		response = json.loads(r.text)

		for entity in response['result']:

			if str(entity['name']) == domain:
				self.cloudFlareIp = entity['content']
				return entity['content']

		return False

	def dump(self):
		r = requests.get('https://api.cloudflare.com/client/v4/zones', headers=self.headers)

		if not r:
			print("Failed to get zones, maybe wrong api key?")
			return False

		response = json.loads(r.text)

		for zone in response['result']:
			print("%s - %s" % (zone['id'], zone['name']))
			rr = requests.get('https://api.cloudflare.com/client/v4/zones/' + zone['id'] + '/dns_records', headers=self.headers)

			response2 = json.loads(rr.text)

			for dns_entry in response2['result']:
				print("\t%s - %s%s" % (dns_entry['id'], dns_entry['name'], ' (proxied)' if dns_entry['proxied'] else ''))

	def setCloudFlareIp(self, ip, zone_id, id, zone_name):
		d = { 'content' : ip, 'zone_id' : zone_id, 'id' : id, 'zone_name' : zone_name, 'type' : 'A', 'name' : 'changeme' }
		r = requests.put('https://api.cloudflare.com/client/v4/zones/' + z + '/dns_records/' + i, headers = self.headers, data = json.dumps(d))

		if not r:
			logWrite("Failed to update record!")
			return False

		response = json.loads(r.text)

		if response['success']:
			return True

		logWrite("Failed, response : %s" % r.text)
		return False

	def resolveIp(self):
		r = requests.get('https://api.ipify.org?format=json', headers=self.headers)

		if not r:
			logWrite("Failed to resolve ip.")
			return False

		response = json.loads(r.text)

		return response['ip']

	def __init__(self, authEmail, authKey):
		self.headers = { 'X-Auth-Email' : authEmail, 'X-Auth-Key' : authKey, 'Content-Type' : 'application/json' }

if u == False or k == False:
	logWrite('Missing api auth tokens.')
	sys.exit(0)

c = CloudFlare(u, k)

if args.dump_mode:
	c.dump();
	sys.exit(0)

cloudflareIp = c.getCloudFlareIp("%s.%s" % (s, d))
publicIp = c.resolveIp()

logWrite('DNS ip %s.' % cloudflareIp)
logWrite('Public ip %s.' % publicIp)

while True:
	if cloudflareIp != publicIp:
		logWrite('Ip has changed to %s, updating.' % publicIp)

		if c.setCloudFlareIp(publicIp, z, i, s):
			cloudflareIp = publicIp
			logWrite('OK.')
		else:
			logWrite('Failed to update.')

	publicIp = c.resolveIp()

	time.sleep(r)


