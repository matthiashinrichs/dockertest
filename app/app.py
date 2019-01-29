from flask import Flask
from flask import render_template
from flask import request
import socket
import os

app = Flask(__name__)
 
 
@app.route('/')
def hello_index():
    os_env = os.environ
    r_ip = request.environ.get('HTTP_X_REAL_IP', request.remote_addr)
    req_env = request.environ
    hostname = socket.gethostname()
    data = {
        'os_env': os_env,
        'client_ip': r_ip,
        'request_env': req_env,
        'hostname': hostname
    }
    return render_template('index.html', title='Home', data=data)


if __name__ == '__main__':
    app.run(debug=True, host='0.0.0.0')
