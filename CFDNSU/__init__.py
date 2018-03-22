#!/usr/bin/env python3

import configparser, argparse, json, pycurl, random, re, os, time, logging, sys
from CFDNSU.Curler import Curler

configPath = '/etc/CFDNSU.conf'
configParserHandler = configparser.ConfigParser()
configParserHandler.read(configPath)
configuration = configParserHandler['configuration']
requiredApiHeaders = ['X-Auth-Email', 'X-Auth-Key']
recordIdentifiers = json.loads(configuration['record_identifiers']) if len(configuration['record_identifiers']) else []

logging.basicConfig(level = logging.INFO, format = '%(asctime)s %(levelname)-8s %(message)s', datefmt = '%Y-%m-%d %H:%M:%S -', handlers = [logging.FileHandler(configuration['log_file']), logging.StreamHandler()])

def authenticate():
	if not os.access(configPath, os.W_OK):
		logging.info('Insufficient permission for writing to %s.' % configPath)
		return False

	for name in requiredApiHeaders:
		current = '[%s]' % configuration[name] if len(configuration[name]) else ''
		value = input('%s%s : ' % (name, current)).strip()

		if value is '' and len(configuration[name]):
			continue

		if value is '':
			logging.info('%s is required, exiting.' % name)

			return False

		configuration[name] = value

	if not verifyAutentication():
		logging.info('Failed to login.')

		return False

	logging.info('Authentication successful, saving configuration.')

	fh = open(configPath, 'w')
	configParserHandler.write(fh)
	fh.close()

	return configParserHandler

def verifyAutentication():
	result = getCurlerHandler()\
		.setOpt(pycurl.URL, 'https://api.cloudflare.com/client/v4/user')\
		.execute()

	result = json.loads(result)

	if result['success']:
		return True

	return False

def removeRecord(fullDomain):
	target = getRecord(fullDomain)

	if not target:
		logging.info('No record for %s found.' % fullDomain)
		return False

	if not os.access(configPath, os.W_OK):
		logging.info('Insufficient permission for writing to %s.' % configPath)
		return False

	recordIdentifiers.remove(target)

	configuration['record_identifiers'] = json.dumps(recordIdentifiers)

	fh = open(configPath, 'w')
	configParserHandler.write(fh)
	fh.close()

def run():
	if not verifyAutentication():
		return False

	if not len(recordIdentifiers):
		logging.info("No record identifiers.")
		return False

	currentIp = getIp()
	if not currentIp:
		return False

	logging.info('Current ip : %s' % currentIp)

	for record in recordIdentifiers:
		ip = getRecordIp(record['fullDomain'])
		logging.info('%s is set to %s' % (record['fullDomain'], ip))

		if ip != currentIp:
			logging.info('Wrong ip for %s updating.' % record['fullDomain'])
			if updateRecord(record, currentIp):
				logging.info('Ok!')

	ipCheck = currentIp

	while True:
		if ipCheck != currentIp:
			currentIp = ipCheck
			logging.info("Ip has changed to %s, updating records." % currentIp)

			if currentIp:
				for record in recordIdentifiers:
					updateRecord(record, currentIp)

		time.sleep(int(configuration['refresh_rate']))
		ipCheck = getIp()

		if not ipCheck:
			ipCheck = currentIp

def getRecordIp(fullDomain):
	record = getRecord(fullDomain)

	result = getCurlerHandler()\
		.setOpt(pycurl.URL, 'https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s' % (record['zoneIdentifier'], record['recordIdentifier']))\
		.execute()

	result = json.loads(result)

	if not result['success']:
		logging.info("Could not get ip.")

	return result['result']['content']

def getRecord(fullDomain):
	for record in recordIdentifiers:
		if record['fullDomain'] == fullDomain:
			return record

	return False

def listRecords():
	logging.info('Listing domain records : ')
	for record in recordIdentifiers:
		logging.info(record['fullDomain'])

def createRecord(fullDomain):
	if getRecord(fullDomain):
		logging.info('Record is already in the list.')
		return False

	if not os.access(configPath, os.W_OK):
		logging.info('Insufficient permission for writing to %s.' % configPath)
		return False

	logging.info('Collecting zone identifiers.')

	zoneIdentifiers = requestZones()

	if not len(zoneIdentifiers):
		logging.info('No active zones found :/')
		return False

	domain = fragmentDomain(fullDomain)
	zoneIdentifier = zoneIdentifiers[domain[1]]
	recordIdentifier = requestRecord(zoneIdentifier, fullDomain)

	recordIdentifiers.append({
		'zoneIdentifier' : zoneIdentifier,
		'recordIdentifier' : recordIdentifier,
		'fullDomain' : fullDomain
	})

	configuration['record_identifiers'] = json.dumps(recordIdentifiers)

	fh = open(configPath, 'w')
	configParserHandler.write(fh)
	fh.close()

	return True

def fragmentDomain(fullDomain):
	result = re.search('(?:([\.\w_0-9]+)\.)?([\.\w_0-9]+\.[\w]{2,}$)', fullDomain, re.IGNORECASE)

	if not result:
		logging.info('%s is not a vaild domain.' % fullDomain)
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

	logging.error(result)

	return False

def requestRecord(zoneIdentifier, fullDomain):
	domainFragment = fragmentDomain(fullDomain)

	if not domainFragment:
		return False

	domain = domainFragment[1]

	dnsRecordRaw = getCurlerHandler()\
		.setOpt(pycurl.URL, 'https://api.cloudflare.com/client/v4/zones/%s/dns_records?per_page=100&type=A' % zoneIdentifier)\
		.execute()

	dnsRecord = json.loads(dnsRecordRaw)

	for record in dnsRecord['result']:
		if record['name'] == fullDomain:
			return record['id']

	return False

def requestZones():
	zoneInformationRaw = getCurlerHandler()\
		.setOpt(pycurl.URL, 'https://api.cloudflare.com/client/v4/zones?per_page=50&sort=name&status=active')\
		.execute()

	zoneInformation = json.loads(zoneInformationRaw)
	domainDictionary = {}
	
	if zoneInformation['success'] == False:
		logging.info('Error failed to get zones.')
		logging.info(zoneInformationRaw)

	for zone in zoneInformation['result']:
		domainDictionary[zone['name']] = zone['id']

	return domainDictionary

def getCurlerHandler():
	for requirement in requiredApiHeaders:
		if configuration[requirement] is '':
			raise Exception('Missing api requirement %s, did you run --authenticate?' % requirement)

	return Curler()\
		.setHTTPHeader({
			'Content-Type' : 'application/json',
			'X-Auth-Email' : configuration['X-Auth-Email'],
			'X-Auth-Key' : configuration['X-Auth-Key']
		})

def getIp():
	ipList = []
	lookupList = [
		'https://ifconfig.co/ip',
		'https://ip.tyk.nu',
		'https://4.ifcfg.me/ip',
		'https://icanhazip.com',
		'https://kekcajwiejaqwqiee.com'
	]

	random.shuffle(lookupList)

	for hostname in lookupList:
		logging.debug('Collecting ip from %s' % hostname)

		try:
			ip = Curler()\
				.setOpt(pycurl.URL, hostname)\
				.setOpt(pycurl.TIMEOUT, 5)\
				.execute()

		except pycurl.error as error:
			logging.warning('Failed to get ip reason(%d) : %s' % (error.args[0], error.args[1]))

		else:
			ip = ip.strip()

			if not ip:
				logging.error('Failed to get the ip.')
				continue

			ipList.append(ip.strip())

			if ipList.count(ip) >= 2:
				return ip

	logging.error('Failed to get ip.')

	return False
