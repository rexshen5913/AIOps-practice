import json
import subprocess
import shlex
from datetime import datetime, timezone
from openai import OpenAI
from kubernetes import client as k8s_client, config as k8s_config

class KubernetesClient:
    def __init__(self, context="alpha"):
        self.context = context
        self.core_v1 = None
        self.app_v1 = None
        self.init_k8s_client()

    def init_k8s_client(self):
        k8s_config.load_kube_config(context=self.context)
        self.core_v1 = k8s_client.CoreV1Api()
        self.app_v1 = k8s_client.AppsV1Api()

    def modify_config(self, service_name, key, value):
        namespace = "default"
        configmap_name = service_name

        try:
            configmap = self.core_v1.read_namespaced_config_map(configmap_name, namespace)
            if configmap.data is None:
                configmap.data = {}
            configmap.data[key] = value
            self.core_v1.patch_namespaced_config_map(configmap_name, namespace, configmap)
            print(f"ConfigMap '{configmap_name}' 已更新")
        except k8s_client.exceptions.ApiException as e:
            print(f"修改 ConfigMap 時發生錯誤: {e}")
            return {"error": str(e)}

        return {"service_name": service_name, "key": key, "value": value}

    def restart_service(self, service_name):
        try:
            deployment = self.app_v1.read_namespaced_deployment(service_name, "default")
            annotations = deployment.spec.template.metadata.annotations or {}
            annotations['kubectl.kubernetes.io/restartedAt'] = datetime.now(timezone.utc).isoformat() + "Z"
            deployment.spec.template.metadata.annotations = annotations
            self.app_v1.patch_namespaced_deployment(service_name, "default", deployment)
        except k8s_client.exceptions.ApiException as e:
            print(f"重新啟動服務時發生錯誤: {e}")
            return {"error": str(e)}

        return {"service_name": service_name, "action": "restart"}
    def apply_manifest(self, resource_type, image):
        command = f"kubectl create {resource_type} {image}-{resource_type} --image={image}"
        try:
            result = subprocess.run(shlex.split(command), capture_output=True, text=True, check=True)
            print(f"{resource_type} {image}-{resource_type} 已建立")
        except subprocess.CalledProcessError as e:
            print(f"命令執行失敗: {e}")
            return {"error": str(e)}

        return {"resource_type": resource_type, "image": image, "result": result.stdout}
class AIOps:
    def __init__(self, api_key, base_url):
        self.client = OpenAI(api_key=api_key, base_url=base_url)
        self.k8s_client = KubernetesClient()

    def handle_tool_call(self, tool_calls, messages):
        available_functions = {
            "modify_config": self.k8s_client.modify_config,
            "restart_service": self.k8s_client.restart_service,
            "apply_manifest": self.k8s_client.apply_manifest,
        }

        for tool_call in tool_calls:
            try:
                function_name = tool_call.function.name
                function_to_call = available_functions[function_name]
                function_args = json.loads(tool_call.function.arguments)
                function_response = function_to_call(**function_args)

                messages.append({
                    "tool_call_id": tool_call.id,
                    "role": "tool",
                    "name": function_name,
                    "content": json.dumps(function_response),
                })
            except Exception as e:
                function_response = {"error": str(e)}
                messages.append({
                    "tool_call_id": tool_call.id,
                    "role": "tool",
                    "name": function_name,
                    "content": json.dumps(function_response),
                })

        return messages

    def run_conversation(self):
        tools = self.get_tools()

        while True:
            query = input("輸入運維指令： (輸入 'exit' 退出): \n")
            if query.lower() == "exit":
                print("結束程序")
                break

            messages = [
                {
                    "role": "system",
                    "content": "你是一個運維專家，能夠根據用戶的需求進行分析，並選擇適當的函數，來執行相應的運維任務。",
                },
                {
                    "role": "user",
                    "content": query,
                },
            ]

            response = self.client.chat.completions.create(
                model="gpt-4o",
                messages=messages,
                tools=tools,
                tool_choice="auto",
            )
            response_message = response.choices[0].message
            tool_calls = response_message.tool_calls

            if tool_calls:
                messages.append(response_message)
                messages = self.handle_tool_call(tool_calls, messages)

                response = self.client.chat.completions.create(
                    model="gpt-4o",
                    messages=messages,
                )

            print(response.choices[0].message.content, "\n\n")

    def get_tools(self):
        return [
            {
                "type": "function",
                "function": {
                    "name": "modify_config",
                    "description": "修改 Kubernetes 中服務的配置，一般配置的名稱是 '服務-config' 的形式，後綴 -config 是固定用法，會以更新配置裡某個鍵所對應的值的方式來進行修改。另外我們在修改後會接著使用 restart_service 函數重新啟動服務",
                    "parameters": {
                        "type": "object",
                        "properties": {
                            "service_name": {"type": "string", "description": "服務的配置名稱，例如：服務是 gateway 的話，配置名稱就是 gateway-config"},
                            "key": {"type": "string", "description": "配置鍵"},
                            "value": {"type": "string", "description": "配置值"},
                        },
                        "required": ["service_name", "key", "value"],
                    },
                },
            },
            {
                "type": "function",
                "function": {
                    "name": "restart_service",
                    "description": "重新啟動服務",
                    "parameters": {
                        "type": "object",
                        "properties": {
                            "service_name": {"type": "string", "description": "kubernetes 中的服務名稱，例如：gateway"},
                        },
                        "required": ["service_name"],
                    },
                },
            },
            {
                "type": "function",
                "function": {
                    "name": "apply_manifest",
                    "description": "部署新的服務",
                    "parameters": {
                        "type": "object",
                        "properties": {
                            "resource_type": {"type": "string", "description": "kubernetes 裡的資源類型，例如：deployment,statefulset"},
                            "image": {"type": "string", "description": "服務的image，例如：nginx像"},
                        },
                        "required": ["resource_type", "image"],
                    },
                },
            },
        ]

aiops = AIOps(api_key="sk-NwaUtVZZdashNZPF68930c47C5Ca41F3Be61D751E37596Cd", base_url="https://api.apiyi.com/v1")
aiops.run_conversation()
