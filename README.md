# GO TCL
 Tcl go binding

 使用CGO实现的GOLANG 执行TCL脚本或命令

# requirements
 tcl8.5 or tcl8.6 
 * tcl.h must in /usr/include
 * tcl8.5.so/tcl8.6.so must in  /usr/lib

# limit
  * Interpreter 执行时有锁,所以不能并行执行
  * 只能指定Tcl脚本，不能自定义Tcl命令


# 参考
  * https://github.com/nsf/gothic.git