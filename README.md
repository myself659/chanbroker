# ChanBroker

# 简单性能测试示例 

## 写文件 
```
root@ubuntu:/share/go/src/github.com/myself659/ChanBroker# time  go  run example/profile.go   -n 2000  -t  60    > profile.txt  
real    1m11.178s
user    0m45.708s
sys     2m13.924s
root@ubuntu:/share/go/src/github.com/myself659/ChanBroker# tail     profile.txt  
0xc42005a1e0 event: 21827
0xc42005a1e0 event: 21828
0xc42005a1e0 event: 21829
0xc42005a1e0 event: 21830
0xc42005a1e0 event: 21831
0xc42005a1e0 event: 21832
0xc42005a1e0 event: 21833
0xc42005a1e0 event: 21834
0xc42005a1e0 has recv: 21835
total: 43670000

```

以real time计算性能结果如下：

每秒发送与接收消息个数： 43670000/72= 606527  
每秒可以发布消息个数：21835/60 = 363 

### 不写文件

```
root@ubuntu:/share/go/src/github.com/myself659/ChanBroker# time  go  run example/profile.go   -n 2000  -t  60  

···
0xc42007c120 has recv: 159444
total: 318888000

real    1m7.561s
user    2m21.824s
sys     0m20.936s
```

以user time计算性能结果如下：

每秒发送与接收消息个数： 318888000/142= 2445690 
每秒可以发布消息个数：159444/60 = 2657   