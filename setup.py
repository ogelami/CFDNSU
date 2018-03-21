from setuptools import setup

setup(name = 'CFDNSU',
	version = '0.3',
	description = 'Cloudflare automatic DNS updater',
	url = 'http://github.com/ogelami/CFDNSU',
	author = 'Robin Dahlberg',
	author_email = 'robin@forwarddevelopment.se',
	license = 'Apache License',
	zip_safe = False,
	packages = ['CFDNSU'],
	scripts = [
		'bin/CFDNSU'
	],
	data_files = [
		('/etc', ['config/CFDNSU.conf'])
	],
	install_requires = [
		'pycurl'
	])
