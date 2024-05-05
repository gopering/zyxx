package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	_ "google.golang.org/grpc/balancer/roundrobin"
	_ "google.golang.org/grpc/resolver/dns" // dns resolver
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"net/http"
	"os"
	"zyxx/pb/manage"
)

var ClientSet = &kubernetes.Clientset{}

func main() {
	WriteConfig()
	InitClient()
	r := gin.Default()
	var app = r.Group("/zyxx")
	app.GET("/ping", func(c *gin.Context) {
		//输出json结果给调用方
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	app.GET("/hello", func(c *gin.Context) {
		fmt.Println("Hello")

		c.JSON(200, gin.H{"hello": "world"})
	})

	r.GET("/job", func(c *gin.Context) {
		JobPod()
		fmt.Println("job successfully")
		c.JSON(200, gin.H{"hello": "world"})
	})
	r.GET("/delete", func(c *gin.Context) {
		DeletePod()
		fmt.Println("delete job successfully")
		c.JSON(200, gin.H{"hello": "world"})
	})

	r.GET("/deleteV2", func(c *gin.Context) {
		DeletePodV2()
		fmt.Println("delete job successfully")
		c.JSON(200, gin.H{"hello": "world"})
	})

	r.GET("/list", func(c *gin.Context) {
		ListPod()
		fmt.Println("delete job successfully")
		c.JSON(200, gin.H{"hello": "world"})
	})
	r.GET("/testPanic", func(c *gin.Context) {
		log.Fatal("Application crashed")
	})

	conn, err := grpc.Dial("manage-svc.mirror.svc.cluster.local:8090", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	manageClient := manage.NewManageClient(conn)

	r.GET("/testRpc", func(c *gin.Context) {
		reply, err := manageClient.Login(context.Background(), &manage.LoginRequest{
			User:     "admin",
			Password: "0192023a7bbd73250516f069df18b500",
		})
		if err != nil {
			log.Fatalf("Could not call RPC function: %v", err)
		}

		c.JSON(http.StatusOK, gin.H{"data": reply})
	})

	//r.Run(":8887")
	// 启动HTTP服务器
	if err := http.ListenAndServe(":8887", r); err != nil {
		panic(err)
	}
}

func InitClient() {
	kubeconfig := "./kube.config"
	// 使用 kubeconfig 文件创建配置加载规则
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.ExplicitPath = kubeconfig

	// 加载配置
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{}).ClientConfig()
	if err != nil {
		fmt.Println("加载配置 clientcmd err==", err.Error())
		log.Fatal(err)
	}

	// 创建 Kubernetes 客户端
	ClientSet, err = kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Println("创建 Kubernetes 客户端 err==", err.Error())
		log.Fatal(err)
	}
}

func JobPod() {
	kubeconfig := "./kube.config"
	// 使用 kubeconfig 文件创建配置加载规则
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.ExplicitPath = kubeconfig

	// 加载配置
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{}).ClientConfig()
	if err != nil {
		fmt.Println("加载配置 clientcmd err==", err.Error())
		log.Fatal(err)
	}

	// 创建 Kubernetes 客户端
	ClientSet, err = kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Println("创建 Kubernetes 客户端 err==", err.Error())
		log.Fatal(err)
	}
	fmt.Println("创建 Kubernetes 客户端 成功")

	fmt.Println("创建 Job 对象 开始")
	// 创建 Job 对象
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: "async-task-job",
		},
		Spec: batchv1.JobSpec{
			Parallelism:             new(int32), // 并行执行的 Pod 的数量
			Completions:             new(int32), // 完成任务的 Pod 的数量
			BackoffLimit:            new(int32), // 失败重试的次数
			TTLSecondsAfterFinished: new(int32), // 完成后保留 Pod 的时间
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Name:  "async-task-container",
							Image: "your-container-image", // 替换为你的容器镜像
							// 添加你的容器的其他配置，例如环境变量、命令等
						},
					},
				},
			},
		},
	}

	// 设置并行执行的 Pod 的数量
	*job.Spec.Parallelism = 3

	// 设置完成任务的 Pod 的数量
	*job.Spec.Completions = 3

	// 设置失败重试的次数
	*job.Spec.BackoffLimit = 3

	// 设置完成后保留 Pod 的时间
	*job.Spec.TTLSecondsAfterFinished = 3600

	// 创建 Job
	createdJob, err := ClientSet.BatchV1().Jobs("mirror").Create(context.TODO(), job, metav1.CreateOptions{})
	if err != nil {
		fmt.Println("创建 Job 对象 error:", err.Error())
		log.Fatal(err)
	}

	fmt.Printf("已创建 Job: %s\n", createdJob.Name)
	fmt.Println("已创建 Job:", createdJob.Name)
}

func DeletePod() {
	/*kubeconfig := "./kube.config"
	// 使用 kubeconfig 文件创建配置加载规则
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.ExplicitPath = kubeconfig

	// 加载配置
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{}).ClientConfig()
	if err != nil {
		fmt.Println("加载配置 clientcmd err==", err.Error())
		log.Fatal(err)
	}

	// 创建 Kubernetes 客户端
	ClientSet, err = kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Println("创建 Kubernetes 客户端 err==", err.Error())
		log.Fatal(err)
	}*/
	err := ClientSet.BatchV1().Jobs("mirror").Delete(context.TODO(), "async-task-job", metav1.DeleteOptions{})
	if err != nil {
		fmt.Println("删除 Job 对象 error:", err.Error())
		log.Fatal(err)
	}
	listOptions := metav1.ListOptions{
		LabelSelector: "app=async-task-job", // 根据标签选择器进行筛选，这里以 "app=my-app" 为例
	}
	podList, err := ClientSet.CoreV1().Pods("mirror").List(context.TODO(), listOptions)
	for _, pod := range podList.Items {
		err := ClientSet.CoreV1().Pods("mirror").Delete(context.TODO(), pod.Name, metav1.DeleteOptions{})
		if err != nil {
			// 处理错误
		}
	}
	fmt.Println("已删除 Job: ", "async-task-job")
}

func DeletePodV2() {
	ClientSetJson, _ := json.Marshal(&ClientSet)
	fmt.Println("打印 ClientSet=", string(ClientSetJson))

	kubeconfig := "./kube.config"
	// 使用 kubeconfig 文件创建配置加载规则
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.ExplicitPath = kubeconfig

	// 加载配置
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{}).ClientConfig()
	if err != nil {
		fmt.Println("加载配置 clientcmd err==", err.Error())
		//log.Fatal(err)
	}

	// 创建 Kubernetes 客户端
	ClientSet, err = kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Println("创建 Kubernetes 客户端 err==", err.Error())
		//log.Fatal(err)
	}
	ClientSetJsons, _ := json.Marshal(&ClientSet)
	fmt.Println("打印 ClientSetJsons=", string(ClientSetJsons))
	err = ClientSet.BatchV1().Jobs("mirror").Delete(context.TODO(), "async-task-job", metav1.DeleteOptions{})
	if err != nil {
		fmt.Println("删除 Job 对象 error:", err.Error())
		//log.Fatal(err)
	}
	listOptions := metav1.ListOptions{
		LabelSelector: "job-name=async-task-job", // 根据标签选择器进行筛选，这里以 "app=my-app" 为例
	}
	podList, err := ClientSet.CoreV1().Pods("mirror").List(context.TODO(), listOptions)
	podListJson, _ := json.Marshal(podList)
	fmt.Println(" pod list : ", string(podListJson))
	for _, pod := range podList.Items {
		err := ClientSet.CoreV1().Pods("mirror").Delete(context.TODO(), pod.Name, metav1.DeleteOptions{})
		if err != nil {
			// 处理错误
			fmt.Println("已删除 pod error: ", err.Error())
		}
		fmt.Println("已删除 pod : ", pod.Name)
	}
	fmt.Println("已删除 Job: ", "async-task-job")
}

func ListPod() {
	ClientSetJson, _ := json.Marshal(&ClientSet)
	fmt.Println("打印 ClientSet=", string(ClientSetJson))

	/*kubeconfig := "./kube.config"
	// 使用 kubeconfig 文件创建配置加载规则
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.ExplicitPath = kubeconfig

	// 加载配置
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{}).ClientConfig()
	if err != nil {
		fmt.Println("加载配置 clientcmd err==", err.Error())
		log.Fatal(err)
	}

	// 创建 Kubernetes 客户端
	ClientSet, err = kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Println("创建 Kubernetes 客户端 err==", err.Error())
		log.Fatal(err)
	}
	ClientSetJsons, _ := json.Marshal(&ClientSet)
	fmt.Println("打印 ClientSetJsons=", string(ClientSetJsons))*/
	/*err := ClientSet.BatchV1().Jobs("mirror").Delete(context.TODO(), "async-task-job", metav1.DeleteOptions{})
	if err != nil {
		fmt.Println("删除 Job 对象 error:", err.Error())
		log.Fatal(err)
	}*/
	listOptions := metav1.ListOptions{
		LabelSelector: "job-name=async-task-job", // 根据标签选择器进行筛选，这里以 "app=my-app" 为例
	}
	podList, err := ClientSet.CoreV1().Pods("mirror").List(context.TODO(), listOptions)
	if err != nil {
		fmt.Println("List pod 对象 error:", err.Error())
		log.Fatal(err)
	}
	podListJson, _ := json.Marshal(podList)
	fmt.Println(" pod list : ", string(podListJson))
	/*	for _, pod := range podList.Items {
			err := ClientSet.CoreV1().Pods("mirror").Delete(context.TODO(), pod.Name, metav1.DeleteOptions{})
			if err != nil {
				// 处理错误
				fmt.Println("已删除 pod error: ", err.Error())
			}
			fmt.Println("已删除 pod : ", pod.Name)
		}
		fmt.Println("已删除 Job: ", "async-task-job")*/

	gracePeriodSeconds := int64(0)
	var backgroundPropagationPolicy = metav1.DeletePropagationBackground
	deleteOptions := metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriodSeconds,
		PropagationPolicy:  &backgroundPropagationPolicy,
	}
	for _, pod := range podList.Items {
		err := ClientSet.CoreV1().Pods("mirror").Delete(context.TODO(), pod.Name, metav1.DeleteOptions{})
		ClientSet.CoreV1().Pods("mirror").Delete(context.TODO(), pod.Name, deleteOptions)
		if err != nil {
			// 处理错误
			fmt.Println("已删除 pod error: ", err.Error())
		}
		fmt.Println("已删除 pod : ", pod.Name)
	}

	isexist, isexistErr := ClientSet.BatchV1().Jobs("mirror").Get(context.TODO(), "async-task-job", metav1.GetOptions{})

	isexistJson, _ := json.Marshal(&isexist)
	fmt.Println("isexistJson  : ", string(isexistJson))
	if isexistErr != nil {
		fmt.Println("isexistErr err : ", isexistErr.Error())
	}

	if isexist.Status.Active == 0 {
		fmt.Println("当前没有活跃的Jobs : ", isexistErr.Error())
	} else {

		err = ClientSet.BatchV1().Jobs("mirror").Delete(context.TODO(), "async-task-job", metav1.DeleteOptions{})
		if err != nil {
			fmt.Println("删除 Job 对象 error:", err.Error())
			log.Fatal(err)
		}
	}
}
func WriteConfig() {

	filePath := "./kube.config"
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Printf("open file error=%v\n", err)
		return
	}
	defer file.Close()
	str := `apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUMvakNDQWVhZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFWTVJNd0VRWURWUVFERXdwcmRXSmwKY201bGRHVnpNQjRYRFRJek1Ea3hOVEV4TURneE9Wb1hEVE16TURreE1qRXhNRGd4T1Zvd0ZURVRNQkVHQTFVRQpBeE1LYTNWaVpYSnVaWFJsY3pDQ0FTSXdEUVlKS29aSWh2Y05BUUVCQlFBRGdnRVBBRENDQVFvQ2dnRUJBTElRClVVTzVBWUgrUW4xR1VNeEZ2U2tPQnNQN2xXWklnLzBDS2VkbWdtLzNWNUVUYkNNVGN0VStmN1djQ0IvOGNLZWEKUW96RSt5cm9kNi96bk9PcisvYTFadkdoS0RBWU9zR3RzWURlaUhBS1BiOHoyU0lPbklYa1p2SFNsZnNFMkFZMgpiSXV1cHRZRzhET2puVzdTekhCanJKTWR4SFVVWnZDbWV1QUwwSnZFSjRtTGRDWDlNM1QwblBqRVpzdlhpMjBMCmpHM2cyc0kxOEtqOXgwUG5qVUo2SUVmbUhUb3BJMHBRSGdhWXNqdVBEOGxSdEYxaUhRWWpHMVJ3S29vdmtJRy8KbjUwWTNrK1U4ZWdxb2R4eURDNmZWUkc0dTY4M1QrbnJKODBhYXpSK2h3ZjhKVVFUZHBSRmFDdU1vODFJT2llKwptWEFoQWN3T2NadjVnSGZ3ZVprQ0F3RUFBYU5aTUZjd0RnWURWUjBQQVFIL0JBUURBZ0trTUE4R0ExVWRFd0VCCi93UUZNQU1CQWY4d0hRWURWUjBPQkJZRUZGUlBhWllRQ3ZEVmU0ZUpWSFZSRHpUWmFsc0hNQlVHQTFVZEVRUU8KTUF5Q0NtdDFZbVZ5Ym1WMFpYTXdEUVlKS29aSWh2Y05BUUVMQlFBRGdnRUJBRE1ZOWNaTmUxeWNITSt2M2g2UAp3cG1zdEhyQ3JCck5qT2VIUFRxam1vcmx2MlZpZjBhY2NlYWR5UThVejJZblgxK1cwOG9aSlJ2SG9rR0xyV0d3CjZtSmc3TlJmYVdCNEN1UTBVOG1MbUIvVWJsc3lUTDBWYmRtMkRheUZ4RnhWMmI1NkZUUWgzK2ZrWmltaEhiaGcKN1VTblFXcENlR0hZbzBIOUR5TGtoV01KSTB0S1VtWWtqTEJ6VU5SMlZZODlaNVE2RmhRdTRYNlRPVUQ3RVlsUwpnR1FERHdLdHg1TDg2aWFpRmN3SVBJRUZFUnFqYXVrTFhDNDB5YVZYQitPUmJwNGtaYmMxUVRoSVBabTV5OWFRCmplQ3B0bHdMcEV1WmxoWlBYaWhJL2cvY05JREhsbHRLd2hhYWZQTEovMFNnOGFZTHFyc3MzazJFU1NqT0JYYXIKc2xvPQotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg==
    server: https://172.21.35.99:6443
  name: kubernetes
contexts:
- context:
    cluster: kubernetes
    user: kubernetes-admin
  name: kubernetes-admin@kubernetes
current-context: kubernetes-admin@kubernetes
kind: Config
preferences: {}
users:
- name: kubernetes-admin
  user:
    client-certificate-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURJVENDQWdtZ0F3SUJBZ0lJYkhOT0hiam1qb2N3RFFZSktvWklodmNOQVFFTEJRQXdGVEVUTUJFR0ExVUUKQXhNS2EzVmlaWEp1WlhSbGN6QWVGdzB5TXpBNU1UVXhNVEE0TVRsYUZ3MHlOREE1TVRReE1UQTRNakJhTURReApGekFWQmdOVkJBb1REbk41YzNSbGJUcHRZWE4wWlhKek1Sa3dGd1lEVlFRREV4QnJkV0psY201bGRHVnpMV0ZrCmJXbHVNSUlCSWpBTkJna3Foa2lHOXcwQkFRRUZBQU9DQVE4QU1JSUJDZ0tDQVFFQTFDeXEyWk1Ed2NXRDhRcXoKSjlLQlM1RVMzc0tLaE4zMW1xVUxRR1NQWnRFMWtNU0VVcVlIcmRKZS9UTjYzaWMzRi9pbjhyWFJvMGZVZDJQQwpmM01VQUFVMTdlMFo3UzVaTEFqeTliclV0a3FZTlZTMUdvcmNBdUMzclhSS2t3UlFuZzNkT1JjQ0txYUtrZHpQCkRKUUdYWkR0ZU1peFZMVXBORVlzL2lsZmdVN21XbVA4SzRmWUd6L0V4U0EzL3FOOWR1eXdqQStQYWljOGRoVXoKWU5nQ1Y2dDA1MVd5K2NndHh6Y2NKMi9jcTNpb1ZkRFZpQitKc2RxSi85S3lRSVNJUElJdmNZTEFwUnUvVUFmbApTRnBNT1VZaXlXU0dNZzN4UHp2RThYUll0ZnZ6VWl3azZCYXpyTk1jMTJMb3F2K3pVQXFUeGtKcjZFbk15YWRNCjM2Y3U2UUlEQVFBQm8xWXdWREFPQmdOVkhROEJBZjhFQkFNQ0JhQXdFd1lEVlIwbEJBd3dDZ1lJS3dZQkJRVUgKQXdJd0RBWURWUjBUQVFIL0JBSXdBREFmQmdOVkhTTUVHREFXZ0JSVVQybVdFQXJ3MVh1SGlWUjFVUTgwMldwYgpCekFOQmdrcWhraUc5dzBCQVFzRkFBT0NBUUVBWlFRRnBUNkhvQ1NVNmwzQzJBTmM1UVlYdlY2d25FanV1WHI4CmVTYlRmWnV4Y2hPS2pvazkycjFSNmdpSjR3S0dxeHhqWndoaUE5NlVmdUlIVkh6R2VPNlVtRVpIZ0RkNGtaaUQKVmJ4Sy9VTHFWSkd3RExGUGRoMGdQMlFtMWp1VTFxQ0lGd0JJS3I0Q0sySzdaSkJLdjVMdmxOMDA4NjFkRE9iYQplRHViL2R5a0tOTGM1a0xFNFpDVERVcEdQZ1BRYi9ncVNCYzdVbndKNnJYYmdUa3BYZ0x2bmRYaVhYMFlSVjdxCkpNWkVYK1Z2YStzR3ArallEaVBMdGJSODE2ZURrQ1ZpbWhmNzg2UTlYVXBwZVk0ejA2SEptNG9lanBzOXFmZEMKd0FWaUs3ZXg1dTJMMzhJZDJESmlZc3N1Mm13cGpZc0JjdE9abU1PWlVLUHZYYnFvQ0E9PQotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg==
    client-key-data: LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFb3dJQkFBS0NBUUVBMUN5cTJaTUR3Y1dEOFFxeko5S0JTNUVTM3NLS2hOMzFtcVVMUUdTUFp0RTFrTVNFClVxWUhyZEplL1RONjNpYzNGL2luOHJYUm8wZlVkMlBDZjNNVUFBVTE3ZTBaN1M1WkxBank5YnJVdGtxWU5WUzEKR29yY0F1QzNyWFJLa3dSUW5nM2RPUmNDS3FhS2tkelBESlFHWFpEdGVNaXhWTFVwTkVZcy9pbGZnVTdtV21QOApLNGZZR3ovRXhTQTMvcU45ZHV5d2pBK1BhaWM4ZGhVellOZ0NWNnQwNTFXeStjZ3R4emNjSjIvY3EzaW9WZERWCmlCK0pzZHFKLzlLeVFJU0lQSUl2Y1lMQXBSdS9VQWZsU0ZwTU9VWWl5V1NHTWczeFB6dkU4WFJZdGZ2elVpd2sKNkJhenJOTWMxMkxvcXYrelVBcVR4a0pyNkVuTXlhZE0zNmN1NlFJREFRQUJBb0lCQUJ1ZzdXYURwRnVaTXNNeQpsMzI2QmFnbmJnT2Y1WlhEcVhYSHhCMVFldlB5amowWnVmbGhNV0xMSUI5c2tyVlcrZStmSzQrSmFaRVBpM1U0CmdsMUNTUHB2czBRV09lZ092d0hpOEhCYk1kUERCdXI2NnRKWC9xcEpST0hMWU9LZ0R6ZGxxc2NDWGYvYnkvT08KZzArcC91STBPeGdkV2lvanBRZnZrM0JOUnVoNWR2WHFTV2RZRVRWdzBUNjE1MDZ6VXpianYyaXY5d2NUVW1vRApVNm1meDFxZS9lakVBajdTUTBGdTRnc0JGQ1ppd09DelZvcHlabkRaT1lGdnVoOXloUi9PZGwrZ2ZKdmZ3eXhsCjRLYzVsS1ZRYmRRanpjaHgzTThYR01GdGlzMmZ6elI0U2cvQzBDdThLby8wWGxmczhCcUcvREdjSDZVKzByTXIKNFduNkR6a0NnWUVBM2x6K2VtZGZNZ285dFBqcy9DVjlLelFSRlR6b1ZYWHRMR3ZLM211MkpIeHdUMlcwV2hMUQpGM3lKbm4wYXNoWVgweFVNa2htcGcrZEljRko1Ui8zL1hGQkZUbmhpR3ZQeUxSVkJ5YWQ3b0FpeTZ5RkY3VVlqClliZG00N3N2ekZqMXBGVUgvK3pYbmg4d0JCcW1ma09ZWCtEdEdPalpxVDEvdEJCc08yWnRVaDhDZ1lFQTlFVWQKRFpEL3A4TFZXRG5scU1nWGtLWmNkN2dMTjVVK1pjc2ZnVVA3bnNvUDM3MDFRL0JmaEVXeHBYSVZXMURXSlZ6MQpQZ25TRnJ1Rm5GTzlRckFOWjE1TklkTk4veUpjeHBuR1lGbVduME9QdEJIbSsrZ0c5ZmxZZEQ2b2ZsbDNiMUZqCnlKWk8zWER6QUdBaW80SzVHVUV2b3dzOGNqdDc3S0NPSk9xU3JmY0NnWUFidWlUUlJvcU1SdEtpK2xjeXFjb04KMVJROFBiZ0swQVdmQUdIdmtpYklMZXdqT2w5ZXkvRyt1L3k5RW9SOXFGdVlLb3ZDdkFoek5pZkdPY0o5dzZKUAo5SUp2NG5yNU9XbjlUU1ZDNit0eWJTTkNSb2ZkcWwxSEZnTnlhaWp2cGpnYklhODVybUxFaU1jSCsvcSt5OWI0ClBhZlM3MVlVMEdKWUphUVpWQkJWcVFLQmdRRER5R0Z3N2piN0QzNVFLSmVhb0VYQytwUkNvSkRkREJIbkpOY3IKbElHbzArdkZPTEhvc2xEY2c3L1BDNUZ5ajJnVXFsMG1URmpIUDZYbmxuYXJiTkJSZVpQNCtKUWJXajlpTHY2QgpXMDBPZWVoRU85VVhNdkhoVk9sQXdyZnFEV3RkSGE4TXB1eXZNRWlVbEhrdTlTZkd4aWlZVmZrczFlQ04yR0lWCjFLMmNJUUtCZ0VnTmd2Qlh0OHUyVVcreHpZcWZtQ04xdCs5dG95aVlCeEI0WDQ5REoxK1ZEa25hT0xqdnFOTXQKTjc0cEFYTUs4bzdOYzFvSVZSeG9mVC9mbVYzaURseVl4b1BSRndFTFkrZXg4WkZpZlhhNlI4RlBMUlRDT3hUTQpzYjNLUUNyWGdHTmFxZ0wxRHdtSXFKdHQ3N1FLdnNMNlhLZXp3eENmQ3c3ZS9GcEZUdUFWCi0tLS0tRU5EIFJTQSBQUklWQVRFIEtFWS0tLS0tCg==
`
	writer := bufio.NewWriter(file)
	writer.WriteString(str)
	writer.Flush()
}
