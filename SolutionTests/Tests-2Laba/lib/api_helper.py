import requests

Base_URL = 'http://127.0.0.1:5000'

def get_page(path=''):
    return requests.get(f'{Base_URL}/{path}')

def post_form(data, path='', json=None):
    return requests.post(f'{Base_URL}/{path}', data=data, json=json)