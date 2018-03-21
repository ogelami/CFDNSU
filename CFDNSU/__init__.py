#!/usr/bin/env python3

import configparser, argparse, json, pycurl, random, re
from CFDNSU.Curler import Curler

configPath = '/etc/CFDNSU.conf'
configParserHandler = configparser.ConfigParser()
configParserHandler.read(configPath)
configuration = configParserHandler['configuration']
requiredApiHeaders = ['X-Auth-Email', 'X-Auth-Key']
recordIdentifiers = json.loads(configuration['record_identifiers']) if len(configuration['record_identifiers']) else []

def authenticate():
	for name in requiredApiHeaders:
		if configuration[name] is '':
			value = input("%s : " % (name)).strip()

			if value is '':
				print("%s is required, exiting." % name)

				return False

			configuration[name] = value

	fh = open(configPath, 'w')
	configParserHandler.write(fh)
	fh.close()	

	return configParserHandler

def createRecord(fullDomain):
	print("Collecting zone identifiers.")

	zoneIdentifiers = requestZones()

	if not len(zoneIdentifiers):
		print('No active zones found :/')
		return False

	domain = fragmentDomain(fullDomain)

	recordIdentifiers.append(createRecord(zoneIdentifiers[domain[1]], fullDomain))

	configuration['record_identifiers'] = json.dumps(recordIdentifiers)

	for record in recordIdentifiers:
		if record['fullDomain'] == fullDomain:
			print('Record is already in the list.')
			return False

	recordId = requestRecord(zoneIdentifier, fullDomain)
	domainFragment = fragmentDomain(fullDomain)

	recordIdentifiers.append({'zoneIdentifier' : zoneIdentifier, 'recordIdentifier' : recordId, 'fullDomain' : fullDomain})

	return True

def fragmentDomain(fullDomain):
	result = re.search('(?:([\.\w_0-9]+)\.)?([\.\w_0-9]+\.[\w]{2,}$)', fullDomain, re.IGNORECASE)

	if not result:
		print("%s is not a vaild domain." % fullDomain)
		return False

	return result.groups()

def updateRecord(record, ip):
	domainFragment = fragmentDomain(record['fullDomain'])

	packet = {
		'type' : 'A',
		'name' : record['fullDomain'],
		'content' : ip
	}

	result = getCurlerHandler()\
		.setOpt(pycurl.URL, 'https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s' % (record['zoneIdentifier'], record['recordIdentifier']))\
		.setOpt(pycurl.CUSTOMREQUEST, 'PUT')\
		.setOpt(pycurl.POSTFIELDS, json.dumps(packet))\
		.execute()

	result = json.loads(result)

	if result['success']:
		return True

	return False

def requestRecord(zoneIdentifier, fullDomain):

	domainFragment = fragmentDomain(fullDomain)

	if not domainFragment:
		return False

	domain = domainFragment[1]

#	'https://api.cloudflare.com/client/v4/zones/' + zone['id'] + '/dns_records'
	dnsRecordRaw = getCurlerHandler()\
		.setOpt(pycurl.URL, 'https://api.cloudflare.com/client/v4/zones/%s/dns_records?per_page=100&type=A' % zoneIdentifier)\
		.execute()

	dnsRecord = json.loads(dnsRecordRaw)

	for record in dnsRecord['result']:
		if record['name'] == fullDomain:
			return record['id']

	return False

def requestZones():
	for requirement in requiredApiHeaders:
		if configuration[requirement] is '':
			print("Cannot dump, missing configuration.")

			return False
	
	zoneInformationRaw = getCurlerHandler()\
		.setOpt(pycurl.URL, 'https://api.cloudflare.com/client/v4/zones?per_page=50&sort=name&status=active')\
		.execute()

	zoneInformation = json.loads(zoneInformationRaw)
	domainDictionary = {}
	
	if zoneInformation['success'] == False:
		print("Error failed to get zones.")
		print(zoneInformationRaw)

	for zone in zoneInformation['result']:
		domainDictionary[zone['name']] = zone['id']

	return domainDictionary

def getCurlerHandler():
	for requirement in requiredApiHeaders:
		if configuration[requirement] is '':
			print("Missing api requirement %s." % requirement)

			return False

	return Curler()\
		.setHTTPHeader({
			'Content-Type' : 'application/json',
			'X-Auth-Email' : configuration['X-Auth-Email'],
			'X-Auth-Key' : configuration['X-Auth-Key']})

def getIp():
	ipList = []
	lookupList = [
		'https://ifconfig.co/ip',
		'https://ip.tyk.nu',
		'https://4.ifcfg.me/ip',
		'https://icanhazip.com',
		'https://kekcajwiejaqwqiee.com']

	random.shuffle(lookupList)

	for hostname in lookupList:
		print('resolving ip from %s' % hostname)

		try:
			ip = Curler()\
				.setOpt(pycurl.URL, hostname)\
				.setOpt(pycurl.TIMEOUT, 5)\
				.execute()

		except pycurl.error as error:
			print("Failed to get ip reason(%d) : %s" % (error.args[0], error.args[1]))

		else:
			ipList.append(ip)

			if ipList.count(ip) >= 2:
				return ip

	print('Failed to get ip.')

	return False