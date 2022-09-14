import numpy as np
import matplotlib.pyplot as plt
import palette
from palette import pc

def main():
    palette.configure(False)
    
    d_indiv, n_indiv = np.loadtxt("indiv_dx_min.txt").T
    d_global, n_global = np.loadtxt("global_dx_min.txt").T
    d_rot_global, n_rot_global = np.loadtxt("global_rotation.txt").T
    d_window_indiv, n_window_indiv = np.loadtxt("rotate_window_indiv.txt").T
    d_mean_indiv, n_mean_indiv = np.loadtxt("rotate_mean_indiv.txt").T
    d_mode_indiv, n_mode_indiv = np.loadtxt("rotate_mode_indiv.txt").T
    
    plt.plot(d_indiv, n_indiv, color=pc("k"),
             label=r"${\rm individual\ (\Delta x)_{\rm min}}$")
    plt.plot(d_global, n_global, color=pc("r"),
             label=r"${\rm global\ (\Delta x)_{\rm min}}$")
    plt.plot(d_rot_global, n_rot_global, color=pc("o"),
             label=r"${\rm global\ rotation}$")
    plt.plot(d_rot_global - 2**16, n_rot_global, color=pc("o"))
    plt.plot(d_window_indiv, n_window_indiv, color=pc("g"),
             label=r"${\rm individual\ rotation\ (window)}$")
    plt.plot(d_window_indiv - 2**16, n_window_indiv, color=pc("g"))
    plt.plot(d_mean_indiv, n_mean_indiv, color=pc("b"),
             label=r"${\rm individual\ rotation\ (mean)}$")
    plt.plot(d_mean_indiv - 2**16, n_mean_indiv, color=pc("b"))
    plt.plot(d_mode_indiv, n_mode_indiv, color=pc("p"),
             label=r"${\rm individual\ rotation\ (mode)}$")
    plt.plot(d_mode_indiv - 2**16, n_mode_indiv, color=pc("p"))

    
    
    plt.legend(loc="upper left", frameon=True, fontsize=16)
    
    xlo, xhi = plt.xlim(-400, 2000)
    ylo, yhi = plt.ylim()
    plt.ylim(ylo, yhi)
    
    plt.xlabel(r"$\Delta x$")
    plt.ylabel(r"$N(\Delta x)$")
    
    n = -512
    while n < xhi:
        plt.plot([n, n], [ylo, yhi], ":", lw=1, c="k")
        n += 256
    
    plt.show()
    
if __name__ == "__main__": main()
