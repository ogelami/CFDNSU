import pycurl
import random

from io import BytesIO

class Curler(object):

	def __init__(self):
		self.curlHandler = pycurl.Curl()
		self.storageIO = BytesIO()
		self.pycurl = pycurl

		self.setOpt(pycurl.WRITEFUNCTION, self.storageIO.write)
		self.setOpt(pycurl.FOLLOWLOCATION, True)
		self.setOpt(pycurl.USERAGENT, self.getRandomUA())

	def getRandomUA(self):
		userAgents = [
			'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/54.0.2840.99 Safari/537.36',
			'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_10_4) AppleWebKit/600.7.12 (KHTML, like Gecko) Version/8.0.7 Safari/600.7.12',
			'Mozilla/5.0 (Windows NT 6.3; Trident/7.0; rv:11.0) like Gecko',
			'Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; Trident/5.0)'
			]

		return random.choice(userAgents)

	def setHTTPHeader(self, header):
		headerResult = []
		
		for key, value in header.items():
			headerResult.append('%s: %s' % (key, value))

		self.setOpt(pycurl.HTTPHEADER, headerResult)

		return self

	def setOpt(self, option, value):
		self.curlHandler.setopt(option, value)

		return self

	def execute(self):
		self.curlHandler.perform()
		self.curlHandler.close()
		
		return self.storageIO.getvalue().decode('UTF-8')