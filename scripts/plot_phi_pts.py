import numpy as np
import matplotlib.pyplot as plt
import palette
from palette import pc

def mvir_to_rvir(mvir):
    return 0.4459216 * (mvir/1e13)**(1.0/3)

def mvir_to_vvir(mvir):
    return 310.6 * (mvir/1e13)**(1.0/3)

def main():
    palette.configure(True)

    file_names = [
        "profiles/phi_pts/phi_pts_4_1.25.txt",
        "profiles/phi_pts/phi_pts_4_2.5.txt",
        "profiles/phi_pts/phi_pts_4_5.txt",
        "profiles/phi_pts/phi_pts_4_10.txt",
        "profiles/phi_pts/phi_pts_4_20.txt"
    ]

    log_r, phi = np.loadtxt(file_names[0]).T
    log_r_2, phi_2 = np.loadtxt(file_names[-1]).T
    
    mvir = 5.112e+09
    vvir = mvir_to_vvir(mvir)
    rvir = mvir_to_rvir(mvir)

    log_r -= np.log10(rvir)
    log_r_2 -= np.log10(rvir)
    
    phi /= vvir**2
    phi_2 /= vvir**2

    plt.plot([-12, 5], [-12, 5], "--", lw=2, c=pc("a"))
    plt.plot(phi, phi_2, ".", c="k")
    
    plt.xlim(-12, 5)
    plt.ylim(-12, 5)
    
    plt.xlabel(r"$\Phi_{\rm true}/V_{\rm vir}^2$")
    plt.ylabel(r"$\Phi(\delta_v = 0.8 V_{\rm vir})/V_{\rm vir}^2$")
    
    plt.fill_between([0, 5], [0, 0], [-12, -12], color=pc("b"), alpha=0.2)
    plt.fill_between([-12, 0], [5, 5], [0, 0], color=pc("r"), alpha=0.2)

    plt.savefig("plots/fig4_vvir_boundedness.png")

if __name__ == "__main__": main()
    
