import pandas as pd
from datetime import datetime
import csv
import matplotlib.pyplot as plt
import matplotlib.dates as mdates
import numpy as np


headers = ['in_size','fanout','oracle_time','Server runtime','Client runtime','VO size','final_amout']
subset_headers = ['in_size', 'fanout', 'subset_size', 'x1', 'y1', 'x2', 'y2', 'Server runtime', 'Client runtime', 'mcs_size', 'sib_size', 'Common time', 'VO size', 'switch']

def PolyCoefficients(x, coeffs):
    """ 
    Taken from: https://stackoverflow.com/a/37352316

    Returns a polynomial for ``x`` values for the ``coeffs`` provided.

    The coefficients must be in ascending order (``x**0`` to ``x**o``).
    """
    o = len(coeffs)
    print(f'# This is a polynomial of order {ord}.')
    y = 0
    for i in range(o):
        y += coeffs[i]*x**i
    return y

def compare_fanouts(df, header, unit, deg):
    f3 = df.query('fanout==3')
    f9 = df.query('fanout==9')
    x1 = f3['in_size']
    y1 = f3[header]
    x2 = f9['in_size']
    y2 = f9[header]
    _, axs = plt.subplots(1, constrained_layout=True)

    z1 = np.polyfit(x1, y1, deg)
    p1 = np.poly1d(z1)
    l1 = np.linspace(0, x1.to_numpy().max(), 1000)

    
    z2 = np.polyfit(x2, y2, deg)
    p2 = np.poly1d(z2)
    l2 = np.linspace(0, x2.to_numpy().max(), 1000)

    axs.plot(x1,y1, 'bo', label='Fanout = 3')
    axs.plot(x2,y2, 'r+', label='Fanout = 9')
    plt.plot(l1, p1(l1), "b", label=str(z1))
    plt.plot(l2, p2(l2), "r", label=str(z1))
    axs.set_xlabel('Input size (# of points)')
    axs.set_ylabel(header+" ("+unit+")")
    plt.grid(True)
    plt.legend()
    plt.savefig(header+"_f3vsf9.eps", format = 'eps')
    plt.savefig(header+"_f3vsf9.png")
    plt.clf()

def compare_areas(df, header, unit, deg):
    large = df.query('y1==50')
    small = df.query('y1==25')
    x1 = large['in_size']
    y1 = large[header]
    x2 = small['in_size']
    y2 = small[header]
    _, axs = plt.subplots(1, constrained_layout=True)

    z1 = np.polyfit(x1, y1, deg)
    p1 = np.poly1d(z1)
    l1 = np.linspace(0, x1.to_numpy().max(), 1000)

    
    z2 = np.polyfit(x2, y2, deg)
    p2 = np.poly1d(z2)
    l2 = np.linspace(0, x2.to_numpy().max(), 1000)

    axs.plot(x1,y1, 'bo', label='Area = 50^2')
    axs.plot(x2,y2, 'r+', label='Area = 25^2')
    plt.plot(l1, p1(l1), "b", label=str(z1))
    plt.plot(l2, p2(l2), "r", label=str(z1))
    axs.set_xlabel('Input size (# of points)')
    axs.set_ylabel(header+" ("+unit+")")
    plt.grid(True)
    plt.legend()
    plt.savefig(header+"_largevssmall.eps", format = 'eps')
    plt.savefig(header+"_largevssmall.png")
    plt.clf()



def create_graph(df, header, unit):

    df = df.query("fanout==3")
    x = df['in_size']
    y = df[header]    
    
    _, axs = plt.subplots(1, constrained_layout=True)

    axs.plot(x,y,'bo')
    axs.set_xlabel('Input size (# of points)')
    axs.set_ylabel(header+" ("+unit+")")
    plt.scatter(x,y)

    plt.grid(True)
    plt.savefig(header+"_f3.eps", format='eps')
    plt.savefig(header+"_"+"f3"+".png")
    plt.clf()



if __name__ == '__main__':
    df = pd.read_csv("1.csv", names = headers)
    df = df.query('oracle_time != 0')
    subdf = pd.read_csv("5.csv", names = subset_headers, nrows = 500)

    subdf['VO size'] = subdf.apply(lambda row: row.mcs_size + row.sib_size, axis=1)

    #center experiments:
    compare_fanouts(df, 'Client runtime', 'ms', 2)
    compare_fanouts(df, 'Server runtime', 'ms', 2)
    compare_fanouts(df, 'VO size', '# of nodes', 2)



    #subset experiments:
    compare_areas(subdf, 'Client runtime', 'ms', 2)
    compare_areas(subdf, 'Server runtime', 'ms', 1)
    compare_areas(subdf, 'VO size', '# of nodes', 1)
    #compare_areas(subdf, 'Common time', 'ms') #sth wrong here


