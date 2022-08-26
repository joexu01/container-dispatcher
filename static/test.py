'''
This is the test code of poisoned training under BadNets.
'''


	import os.path as osp


import cv2
from sklearn.utils import shuffle
import torch
import torch.nn as nn
import torchvision
from torchvision.datasets import DatasetFolder
from torchvision.transforms import Compose, ToTensor, RandomHorizontalFlip, ToPILImage, Resize

import core
import pdb
import numpy as np


for i in range(1):

	# ========== Set global settings ==========
	global_seed = i
	deterministic = True
	torch.manual_seed(global_seed)
	CUDA_VISIBLE_DEVICES = '1'
	datasets_root_dir = './datasets'


	# ========== BaselineMNISTNetwork_MNIST_BadNets ==========
	dataset = torchvision.datasets.MNIST

	transform_train = Compose([
    ToTensor()
])
	trainset = dataset(datasets_root_dir, train=True, transform=transform_train, download=True)

	transform_test = Compose([
    ToTensor()
])
	testset = dataset(datasets_root_dir, train=False, transform=transform_test, download=True)

	pattern = torch.zeros((28, 28), dtype=torch.uint8)
	pattern[-3:, -3:] = 255
	weight = torch.zeros((28, 28), dtype=torch.float32)
	weight[-3:, -3:] = 1.0

	badnets = core.BadNets(
	    train_dataset=trainset,
	    test_dataset=testset,
	    model=core.models.BaselineMNISTNetwork(),
	    loss=nn.CrossEntropyLoss(),
	    y_target=1,
	    poisoned_rate=0.05,
	    pattern=pattern,
	    weight=weight,
	    seed=global_seed,
	    deterministic=deterministic
	)

	# Train Attacked Model (schedule is set by yamengxi)
	schedule = {
	    'device': 'GPU',
	    'CUDA_VISIBLE_DEVICES': CUDA_VISIBLE_DEVICES,
	    'GPU_num': 1,

	    'benign_training': True,
	    'batch_size': 128,
	    'num_workers': 2,

	    'lr': 0.1,
	    'momentum': 0.9,
	    'weight_decay': 5e-4,
	    'gamma': 0.1,
	    'schedule': [150, 180],

	    'epochs': 10,

	    'log_iteration_interval': 100,
	    'test_epoch_interval': 10,
	    'save_epoch_interval': 10,

	    'save_dir': 'experiments',
	    'experiment_name': 'BaselineMNISTNetwork_MNIST_BadNets'
	}
	#badnets.train(schedule)
	#torch.save(badnets,'/home/yixiao/xyx/backdoor_detection/BackdoorBox/pretrained_models/minist/clean/'+str(i)+'.pth')
#pdb.set_trace()

value_box = []
mean_value = 0
true_pos = 0
for ii in range(100):
	model=torch.load('/home/yixiao/xyx/backdoor_detection/BackdoorBox/pretrained_models/minist/badnet/'+str(ii)+'.pth','cpu')
	model=model.model
	model.eval()
	logits = torch.zeros(10)


	test_images = torch.zeros(10,32,28,28)

	for target_class in range(10):
		count = 0
		for i in range(10000):
			if trainset.targets[i] == target_class:
				test_images[target_class,count] = trainset.data[i].clone()
				count += 1
			if count == 32:break

	for i in range(10):
		aa = torch.sum(model(test_images[i].unsqueeze(1).to(torch.float32)),0)
		aa[torch.argmax(aa)]=0
		logits += aa

	#pdb.set_trace()
	value_box.append(torch.max(logits).detach().numpy()/32)

print(np.sort(value_box))
np.savetxt('0.txt',np.sort(value_box))
pdb.set_trace()