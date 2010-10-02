# Copyright 2010 Fumitoshi Ukai
# Use of this source code is governed by a BSD-style

# gomake でビルド
# gomake install でビルドしたパッケージをインストールできます。

include $(GOROOT)/src/Make.inc

# TARGにパッケージのインポートパスを設定します。
TARG=godentaku.googlecode.com/hg/godentaku
# GOFILESにパッケージのソースファイル一式を設定します。
GOFILES=\
	godentaku.go\

# パッケージの場合 Make.pkgをincludeします。
include $(GOROOT)/src/Make.pkg
