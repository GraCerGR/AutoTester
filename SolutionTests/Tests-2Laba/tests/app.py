from flask import Flask, request, abort
from tests.Solution import Solution

app = Flask(__name__)
sut = Solution()

@app.get('/')
def form():
    return '<html><body><h1>Уникальные пути</h1>' \
           '<form action="/" method="POST">' \
           '<input type="text" name="m">' \
           '<input type="text" name="n">' \
           '<input type="submit">' \
           '</form></body></html>'

@app.post('/')
def unicPaths():
    try:
        m = int(request.form['m'])
        n = int(request.form['n'])
        return {"result": sut.uniquePaths(m, n)}
    except TypeError:
        abort(400)

    except ValueError:
        abort(400)


if __name__ == '__main__':
    app.run(debug=True)