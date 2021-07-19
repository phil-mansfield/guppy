import numpy as np
import matplotlib.pyplot as plt
import palette
from palette import pc
import sys

deltas = ["1.25", "2.5", "5", "10", "20"]
colors = [pc("k"), pc("r"), pc("o"), pc("b"), pc("p")]

def mvir_to_rvir(mvir):
    return 0.4459216 * (mvir/1e13)**(1.0/3)

def mvir_to_vvir(mvir):
    return 310.6 * (mvir/1e13)**(1.0/3)

def plot_rho(i_low, i_high):
    for i in range(i_low, i_high+1):
        plt.figure()
        
        for j, delta in enumerate(deltas):
            fname = "profiles/delta_%s_1.25/rho.%d.txt" % (delta, i)
            r, rho_r2 = np.loadtxt(fname).T

            plt.plot(r, rho_r2, c=colors[j],
                     label=r"$\delta_x = %s\ h^{-1}{\rm kpc}, \delta_v = %s\ {\rm km/s}$" % (delta, delta))

        plt.xscale("log")
        plt.yscale("log")
        plt.xlabel(r"$r\ (h^{-1}{\rm Mpc})$")
        plt.ylabel(r"$(\rho/\bar{\rho})\,r\ (h^{-1}{\rm Mpc})$")
        plt.legend(loc="lower right", fontsize=16)

        y_low, y_high = plt.ylim()
        plt.ylim(y_low, y_high)
        x_low, x_high = plt.xlim()
        plt.xlim(x_low, x_high)
    
        plt.fill_between([x_low, 4e-3], [y_low, y_low], [y_high, y_high],
                         color="k", alpha=0.2)

def MA21_bias(r, epsilon):
    h = epsilon / 0.357
    A = 0.172
    beta = -0.522
    return 1 - np.exp(-(A*h/r)**beta)
        
def plot_vcirc(i_low, i_high, mvir):
    for i in range(i_low, i_high+1):
        plt.figure()

        mvir_i = mvir[i]
        rvir_i = mvir_to_rvir(mvir_i)
        vvir_i = mvir_to_vvir(mvir_i)
        
        vcirc_base = None

        alpha = -2.0/3
        
        for j, delta in enumerate(deltas):
            fname = "profiles/delta_%s_1.25/vcirc.%d.txt" % (delta, i)
            r, vcirc = np.loadtxt(fname).T

            if vcirc_base is None: vcirc_base = vcirc

            if j == 0:
                plt.plot(r, (vcirc/vvir_i)*(r/rvir_i)**alpha, c=colors[j])
            else:
                plt.plot(r, (vcirc/vvir_i)*(r/rvir_i)**alpha, c=colors[j],
                         label=(r"$\delta_x=%s\ h^{-1}{\rm kpc}=%.3f\cdot l$" %
                                (delta, float(delta) / (62.5e3 / 1024))))
            
            if j == 0:
                bias = MA21_bias(r, 1e-3)
                high = (vcirc/vvir_i)*bias*(r/rvir_i)**alpha
                low = (vcirc/vvir_i)/bias*(r/rvir_i)**alpha
                plt.fill_between(r, low, high, alpha=0.2, color="k")

                
        plt.xscale("log")
        plt.ylim(1, 20)
        plt.yscale("log")
        plt.xlabel(r"$r\ (h^{-1}{\rm Mpc})$")
        plt.ylabel(r"$(V_{\rm circ}(<r)/V_{\rm vir})\cdot(r/R_{\rm vir})^{-2/3}$")
        #plt.legend(loc="lower left", fontsize=16, frameon=True)

        y_low, y_high = plt.ylim()
        plt.ylim(y_low, y_high)
        x_low, x_high = plt.xlim()
        plt.xlim(x_low, x_high)
    
def plot_shape(i_low, i_high, mvir):
    for i in range(i_low, i_high+1):
        plt.figure()
        
        for j, delta in enumerate(deltas):
            fname = "profiles/delta_%s_1.25/shape.%d.txt" % (delta, i)
            r, c_to_a, b_to_a = np.loadtxt(fname).T
            ok = c_to_a > 0

            if j == 0:
                plt.plot(r[ok], c_to_a[ok], c=colors[j])
            else:
                plt.plot(r[ok], c_to_a[ok], c=colors[j],
                         label=(r"$\delta_x=%s\ h^{-1}{\rm kpc}=%.3f\cdot l$" %
                                (delta, float(delta) / (62.5e3 / 1024))))

                
        plt.xscale("log")
        plt.xlabel(r"$r\ (h^{-1}{\rm Mpc})$")
        plt.ylabel(r"$(c/a)(r)$")
        plt.legend(loc="lower right", fontsize=16)

        y_low, y_high = plt.ylim()
        plt.ylim(y_low, y_high)
        x_low, x_high = plt.xlim()
        plt.xlim(x_low, x_high)
    
        plt.fill_between([x_low, 4e-3], [y_low, y_low], [y_high, y_high],
                         color="k", alpha=0.2)

def plot_l(i_low, i_high, mvir):
    for i in range(i_low, i_high+1):
        plt.figure()

        mvir_i = mvir[i]
        vvir_i = mvir_to_vvir(mvir_i)
        
        for j, delta in enumerate(deltas):
            fname = "profiles/delta_1.25_%s/l_bullock.%d.txt" % (delta, i)
            r, l_bullock = np.loadtxt(fname).T
            ok = l_bullock > 0

            if j == 0:
                plt.plot(r[ok], l_bullock[ok], c=colors[j])
            else:
                plt.plot(r[ok], l_bullock[ok], c=colors[j],
                         label=((r"$\delta_v = %s\,{\rm km\,s^{-1}} " +
                                 r"= %.3f\,V_{\rm vir}$") %
                                (delta, float(delta) / vvir_i)))
                
        plt.xscale("log")
        plt.yscale("log")
        plt.xlabel(r"$r\ (h^{-1}{\rm Mpc})$")
        plt.ylabel(r"$|\vec{L}(r)|/(M(r)V_{\rm circ}(r)r)$")

        y_low, y_high = plt.ylim()
        plt.ylim(y_low, y_high)
        x_low, x_high = plt.xlim()
        plt.xlim(x_low, x_high)
    
        plt.fill_between([x_low, 4e-3], [y_low, y_low], [y_high, y_high],
                         color="k", alpha=0.2)

def plot_f_bound(i_low, i_high, mvir):
    for i in range(i_low, i_high+1):
        plt.figure()

        mvir_i = mvir[i]
        vvir_i = mvir_to_vvir(mvir_i)
        
        for j, delta in enumerate(deltas):
            fname = "profiles/delta_1.25_%s/f_bound.%d.txt" % (delta, i)
            r, f_bound = np.loadtxt(fname).T
            ok = f_bound > 0

            if j == 0:
                plt.plot(r[ok], f_bound[ok], c=colors[j])
            else:
                plt.plot(r[ok], f_bound[ok], c=colors[j],
                         label=((r"$\delta_v = %s\,{\rm km\,s^{-1}} " +
                                 r"= %.3f\,V_{\rm vir}$") %
                                (delta, float(delta) / vvir_i)))

        plt.xscale("log")
        plt.xlabel(r"$r\ (h^{-1}{\rm Mpc})$")
        plt.ylabel(r"$f_{\rm bound}(r)$")
        plt.legend(loc="lower left", fontsize=16, frameon=True)

        y_low, y_high = plt.ylim()
        plt.ylim(y_low, y_high)
        x_low, x_high = plt.xlim()
        plt.xlim(x_low, x_high)
    
        plt.fill_between([x_low, 4e-3], [y_low, y_low], [y_high, y_high],
                         color="k", alpha=0.2)

        
        
def main():
    palette.configure(True)

    mvir = np.loadtxt("profiles/target_haloes.txt", usecols=(0,))
    low, high = int(sys.argv[1]), int(sys.argv[2])
    
    #plot_vcirc(low, high, mvir)
    #plt.savefig("plots/fig_3a_vcirc.png")
    #plot_shape(low, high, mvir)
    #plt.savefig("plots/fig_3b_shape.png")
    plot_l(low, high, mvir)
    plt.savefig("plots/fig_3c_lambda.png")
    plot_f_bound(low, high, mvir)
    plt.savefig("plots/fig_3d_f_bound.png")

    #plt.show()

if __name__ == "__main__": main()
