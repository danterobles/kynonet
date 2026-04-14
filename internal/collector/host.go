package collector

import (
	"context"
	"fmt"

	"github.com/denisbrodbeck/machineid"
	"github.com/shirou/gopsutil/v3/host"
)

// collectHost recopila la información de identidad del equipo:
// nombre de host, sistema operativo, plataforma, arquitectura del kernel
// y un UUID estable del hardware.
//
// El machine ID se obtiene de fuentes dependientes del SO:
//   - Linux: /etc/machine-id
//   - Windows: clave de registro MachineGuid
//   - macOS: IOPlatformSerialNumber vía ioreg
//
// Si la lectura del machine ID falla (por ejemplo en contenedores sin /etc/machine-id),
// el campo se deja vacío en lugar de abortar la recolección completa.
func collectHost(ctx context.Context) (HostInfo, error) {
	info, err := host.InfoWithContext(ctx)
	if err != nil {
		return HostInfo{}, fmt.Errorf("host info: %w", err)
	}

	mid, err := machineid.ID()
	if err != nil {
		// Error no fatal: el machine ID puede no estar disponible en
		// contenedores o entornos con permisos restringidos.
		mid = ""
	}

	return HostInfo{
		Hostname:  info.Hostname,
		OS:        info.OS,
		Platform:  info.Platform,
		Arch:      info.KernelArch, // arquitectura real del kernel, no la del binario compilado
		MachineID: mid,
	}, nil
}
