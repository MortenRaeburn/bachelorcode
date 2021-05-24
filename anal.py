import pandas as pd
from datetime import datetime
import csv
import matplotlib.pyplot as plt
import matplotlib.dates as mdates
import numpy as np


headers = ['in_size','fanout','oracle_time','Server runtime','Client runtime','VO size','final_amout']
subset_headers = ['in_size', 'fanout', 'subset_size', 'x1', 'y1', 'x2', 'y2', 'Server runtime', 'Client runtime', 'mcs_size', 'sib_size', 'Common time', 'VO size']

def compare_fanouts(df, header, unit):
    f3 = df.query('fanout==3')
    f9 = df.query('fanout==9')
    x1 = f3['in_size']
    y1 = f3[header]
    x2 = f9['in_size']
    y2 = f9[header]
    _, axs = plt.subplots(1, constrained_layout=True)

    axs.plot(x1,y1, 'bo', label='Fanout = 3')
    axs.plot(x2,y2, 'r+', label='Fanout = 9')
    axs.set_xlabel('Input size (# of points)')
    axs.set_ylabel(header+" ("+unit+")")
    plt.grid(True)
    plt.legend()
    #plt.savefig(header+"_f3vsf9.eps", format = 'eps')
    plt.savefig(header+"_f3vsf9.png")
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
    #plt.savefig(header+"_f3.eps", format='eps')
    plt.savefig(header+"_"+"f3"+".png")
    plt.clf()



if __name__ == '__main__':
    df = pd.read_csv("test2res1.csv", names = headers)
    subdf = pd.read_csv("subtest.csv", names = subset_headers)

    subdf['VO size'] = subdf.apply(lambda row: row.mcs_size + row.sib_size, axis=1)

    #center experiments:
    compare_fanouts(df, 'Client runtime', 'ms')
    compare_fanouts(df, 'Server runtime', 'ms')
    compare_fanouts(df, 'VO size', '# of nodes')



    #subset experiments:
    create_graph(subdf, 'Client runtime', 'ms')
    create_graph(subdf, 'Server runtime', 'ms')
    create_graph(subdf, 'VO size', '# of nodes')
    create_graph(subdf, 'Common time', 'ms')


