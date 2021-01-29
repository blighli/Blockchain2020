A simple PBFT protocol over HTTP, using python3 asyncio/aiohttp
A pbft.yaml config file is needed for the nodes to run. A sample is presented in the file.
------------------------------------------------------------------
Run the nodes:
for i in {0..3}; do python ./node.py -i $i -lf False &; done
--------------------------------------------------------------------
Send request to any one of the nodes:
curl -vLX POST --data '{ 'id': (0, 0), 'client_url': http://localhost:20001/reply, 'timestamp': time.time(),'data': 'data_string'}' http://localhost:30000/request 

id is a tuple of (client_id, seq_id)
client_url is the url for sending request to the get_reply function
timestamp is the current time
data is whatever data in string format
---------------------------------------------------------------------
Run the clients
for i in {0...2}; do python client.py -id $i -nm 5 &; done