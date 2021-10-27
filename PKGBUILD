# Maintainer: myyc

pkgname=clockhead-git
_pkgname=clockhead
pkgver=r3.e2dcd27
pkgrel=1
pkgdesc="A CPU frequency scaling daemon for Linux with no configuration."
arch=("any")
url="https://github.com/myyc/clockhead"
license=('BSD')
makedepends=('go')
provides=("$_pkgname")
conflicts=("$_pkgname")
source=("${_pkgname}::git+https://github.com/myyc/${_pkgname}.git")
sha256sums=('SKIP')

pkgver() {
  	cd "${srcdir}/${_pkgname}"

	printf "r%s.%s" "$(git rev-list --count HEAD)" "$(git rev-parse --short HEAD)"
}

build() {
  	cd "${srcdir}/${_pkgname}"

  	go build *.go
}

package() {
	install -Dm700 "${srcdir}/${_pkgname}/clockhead" "${pkgdir}/usr/bin/clockhead"
	install -Dm644 "${srcdir}/${_pkgname}/LICENSE" "${pkgdir}/usr/share/licenses/$pkgname/LICENSE"
	install -Dm644 "${srcdir}/${_pkgname}/README.md" "${pkgdir}/usr/share/doc/$pkgname/README"
	install -Dm644 "${srcdir}/${_pkgname}/${_pkgname}.service" "${pkgdir}/usr/lib/systemd/system/${_pkgname}.service"
}
