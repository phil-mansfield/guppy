import numpy as np
import matplotlib.pyplot as plt
import palette
from palette import pc

sim_names = ["Erebos_CBol_L63"]
accuracies = [
    [1.25, 2.5, 5, 10, 20]
]
Ls = [62.5]

colors = [pc("r"), pc("o"), pc("g"), pc("b"), pc("p")]


def snap_lookup_tables(file_name):
    snap, z, a = np.loadtxt(file_name).T
    z[63] = 10**((np.log10(z[62]) + np.log10(z[64])) / 2)
    a[63] = 10**((np.log10(a[62]) + np.log10(a[64])) / 2)
    return z, a
    
def main():
    snap_to_z, snap_to_a = snap_lookup_tables("profiles/redshifts.txt")
    
    L, sim_name = Ls[0], sim_names[0]
    dx_means = [0]*len(accuracies[0])
    dv_means = [0]*len(accuracies[0])

    spacing = L / 1024
    
    for i, acc in enumerate(accuracies[0]):
        str_acc = str(acc)
        dx_file_name = "profiles/sizes/%s/dx_%s.txt" % (sim_name, str_acc)
        dv_file_name = "profiles/sizes/%s/dv_%s.txt" % (sim_name, str_acc)

        snap = np.arange(101)
        snap = snap[snap != 63]

        dx_size = np.loadtxt(dx_file_name, usecols=(0,))[:-1]
        dv_size = np.loadtxt(dv_file_name, usecols=(0,))[:-1]
        dx_size = dx_size[dx_size != 0]
        dv_size = dv_size[dv_size != 0]

        exp_small = 1024**3 * (4 + 12) / 4**3 / 2**10
        exp_large = 1024**3 * (4 + 12) / 2**10
        
        dx_ratio = np.zeros(len(dx_size))
        dx_ratio[:-1] = dx_size[:-1] / exp_small
        dx_ratio[-1] = dx_size[-1] / exp_large
        
        dv_ratio = np.zeros(len(dv_size))
        dv_ratio[:-1] = dv_size[:-1] / exp_small
        dv_ratio[-1] = dv_size[-1] / exp_large
        
        dx_means[i], dv_means[i] = np.mean(dx_ratio), np.mean(dv_ratio)
        
        if i == 1:
            plt.plot(snap_to_a[snap], dx_ratio, c=pc("r"),
                     label=r"$\delta x = %.3f\,l$" % (acc/spacing/1e3))
            plt.plot([1/21, 1], [dx_means[i], dx_means[i]], "--",
                     lw=2, c=pc("r"))
        if i == 3:
            plt.plot(snap_to_a[snap], dv_ratio, c=pc("b"),
                     label=r"$\delta v = %s\,{\rm km\,s^{-1}}$" % str_acc)
            plt.plot([1/21, 1], [dx_means[i], dx_means[i]], "--",
                     lw=2, c=pc("b"))


    plt.legend(loc="upper left", fontsize=16)
    plt.xscale("log")
    plt.ylabel(r"${\rm Compression\ ratio}$")
    plt.xlabel(r"$a(z)$")
    plt.xlim(1/21, 1)

    plt.savefig("plots/fig_2a_ratio_vs_a.png")

    plt.figure()
    
    acc = np.array(accuracies[0])
    
    plt.plot(np.log10(acc), dv_means, c=pc("b"), label=r"$\vec{v}$")
    plt.plot(np.log10(acc), dv_means, "o", c=pc("b"))
    plt.xlabel(r"$\log_{10}(\delta v)\ ({\rm km\,s^{-1}})$", color=pc("b"))
    plt.ylabel(r"${\rm Compression\ ratio}$")
    plt.tick_params(axis="x", colors=pc("b"))

    plt.twiny()
    
    plt.plot(np.log10(acc/spacing/1e3), dx_means, c=pc("r"), label=r"$\vec{x}$")
    plt.plot(np.log10(acc/spacing/1e3), dx_means, "o", c=pc("r"))
    plt.xlabel(r"$\log_{10}(\delta x / l)$", color=pc("r"))
    plt.ylabel(r"${\rm Compression\ ratio}$")
    plt.tick_params(axis="x", colors=pc("r"))

    plt.savefig("plots/fig_2b_ratio_vs_accuracy.png")
    
if __name__ == "__main__":
    palette.configure(True)
    
    main()

    #plt.show()
