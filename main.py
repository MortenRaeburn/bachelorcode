#!/usr/bin/python3

import time

from Centerpoint import *
from flask import Flask, redirect, url_for, request
import json

app = Flask(__name__)


@app.route('/centerpoint', methods=['POST'])
def centerpoint():
    if request.method == 'POST':
        ps = request.json

        point_set = []

        for p in ps:
            point_set.append(Point(p[0], p[1]))

        plot = False
        start_time = time.time()
        cp = Centerpoint(point_set, plot=plot)
        centerpoint = cp.reduce_then_get_centerpoint()
        # centerpoint = cp.brute_force_centerpoint()
        print("Total time used for %d points is: %.2f s" % (100, time.time() - start_time))  
        # plotResult(point_set, centerpoint)

        res = centerpoint

        res_json = json.dumps(res, default=lambda o: o.__dict__, indent=4, sort_keys=True)

        # print(res_json)

        return res_json
  


def plotResult(point_set, cp):
    plt.clf()
    x_min, x_max = find_x_bounds(point_set)
    interval = Interval(x_min - 10, x_max + 10)
    y_min, y_max = find_y_bounds(point_set)
    prepare_axis(interval.l - 5, interval.r + 5, y_min - 5, y_max + 5)
    plot_point_set(point_set, color='b')
    plot_point(cp, color='r')
    plt.pause(1)
    end = input('Press enter to the next step.')


if __name__ == '__main__':
    # app.debug = True
    app.run() #go to http://localhost:5000/ to view the page.
p