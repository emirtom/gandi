import requests

class GandiClient:
    def __init__(self, uri, api_key="default") -> None:
        self.__URI = uri
        self.__api_key = api_key
        self.header = {'Content-Type': 'application/json',}
    
    def create_collection(self, collection_name = "default", dim=128) -> None:
        data = {
            "collectionName": collection_name,
            "dimension": dim,
        }
        res = requests.post(
            'http://' + self.__URI + "/gandi/collections/create",
            headers=self.header,
            json=data
        )
        
        res = res.json()
        
        if res["code"] == 200:
            print("Collection succesfully created")
        else:
            print("Error in creating collection")
    
    
    def insert(self, collection_name="default", data=[]):
        in_data = {
            'data': data,
            'collectionName': collection_name
        }
        
        res = requests.post(
            'http://' + self.__URI + '/gandi/entities/insert',
            headers=self.header,
            json=in_data
        )
        
        res = res.json()
        
        
        if res["code"] == 200:
            print("Insert successful")
        else:
            print("Insert failed")
    
    
        
    def get(self, collection_name="default", ids=[]):
        data = {
            "collectionName": collection_name,
            'id': ids,
        }
        
        
        res = requests.post(
            'http://' + self.__URI + '/gandi/entities/get',
            headers=self.header,
            json=data
        )
        
        res = res.json()
        
        if res["code"] == 200:
            print(res["data"])
        else:
            print("Get failed")
            
            
client = GandiClient("localhost:19530")

client.get()