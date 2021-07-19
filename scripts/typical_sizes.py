import numpy as np
import matplotlib.pyplot as plt
import palette
from palette import pc

target = "profiles/sizes_full/dx_%s_%03d.txt"

acc_strs = ["1.25", "2.5", "5"]
snaps = [0, 20, 40, 60, 80, 100]

colors = [pc("k"), pc("p"), pc("b"), pc("g"), pc("o"), pc("r")]

def vec_to_idx(x, y, z):
    return x + y*8 + z*64 

def plot_dev(i_acc):
    ratios = np.zeros((6, 7*7*7))
    acc = acc_strs[i_acc]

    plt.figure(i_acc)
    
    for i_snap, snap in enumerate(snaps):
        color = colors[i_snap]
        sizes = np.loadtxt(target % (acc, snap), usecols=(0,))
        names = np.loadtxt(target % (acc, snap), usecols=(1,), dtype=str)
        sizes = np.reshape(sizes, (8, 8, 8))
        names = np.reshape(names, (8, 8, 8))

        mean_size = np.mean(sizes)
        
        block_sizes = ratios[i_snap, :]
        i = 0

        for ox in range(7):
            for oy in range(7):
                for oz in range(7):
                    block = sizes[ox:ox+2, oy:oy+2, oz:oz+2]
                    block_sizes[i] = np.mean(block)
                    i += 1

        block_sizes /= mean_size
        plt.plot(block_sizes, c=color, lw=1.5)
    
    x_low, x_high = plt.xlim()
    plt.xlim(x_low, x_high)

    dev = np.sqrt(np.sum((ratios-1)**2, axis=0)/6)
    
    plt.plot([x_low, x_high], [1, 1], "--", lw=2, c="k")
    
    plt.figure(4)

    color = [pc("r"), pc("o"), pc("b")][i_acc]
    
    plt.plot(dev, c=color, lw=1.5)
    plt.yscale("log")

def main():
    palette.configure(False)

    print("Good target: center on %d%d%d" % (47 % 7, (47 // 7) % 7, 47 // 49))
    
    plot_dev(0)
    plot_dev(1)
    plot_dev(2)

    plt.show()

if __name__ == "__main__": main()
