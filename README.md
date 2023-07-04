# GO TCL
 Tcl go binding

 使用CGO实现的GOLANG 执行TCL脚本或命令

# requirements
 tcl8.5 or tcl8.6 
 * tcl.h must in /usr/include
 * tcl8.5.so/tcl8.6.so must in  /usr/lib

# 限制
  * Interpreter 执行时有锁,所以不能并行执行
  * 只能指定Tcl脚本，不能自定义Tcl命令
# 注意
  * 如果在tcl解析器中调用exit命令会导致整个程序退出
  * 如果在tcl解析器中创建了子进程，在关闭解析器对象时，`不会`关闭这些子进程

# 参考
  * https://github.com/nsf/gothic.git
