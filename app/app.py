from flask import Flask
import socket 

app = Flask(__name__)
 
 
@app.route('/')
def hello_whale():
    hostname = socket.gethostname()
    return "Whale, Hello there! %s"%(hostname)
 
if __name__ == '__main__':
    app.run(debug=True, host='0.0.0.0')
