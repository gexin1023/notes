## 1. Two Sum

+ 输入：一个整数数组nums[]，一个整数target
+ 返回：数组中找出2个数，使其和为target，并返回两个数在数列中的索引。

比如:

    Given nums = [2, 7, 11, 15], target = 9,

    Because nums[0] + nums[1] = 2 + 7 = 9,
    return [0, 1].

**方法1：**
最简单的遍历方法，时间复杂度O(n^2)，空间复杂度O(1)，代码如下：

	class Solution {
	public:
    	vector<int> twoSum(vector<int>& nums, int target) {
    	
        for(int i = 0 ; i< nums.size(); i++)
            for(int j = i+1; j<nums.size(); j++)
                if((nums[i] + nums[j]) == target)
                    return vector<int>{i,j};
        
    	}
	};

**方法2：**
将数组中元素加入HashMap中，然后遍历一遍，找出对应的元素。时间复杂度为O(n)，空间复杂度O(n)。

**下边代码有误，我需要先学习下C++STL再完善，STL遗忘了。。。**

	class Solution {
	public:
    		vector<int> twoSum(vector<int>& nums, int target) {
    	
	    		Map<Integer, Integer> map = new hashmap<>();
	    		for(int i = 0; i< nums.size(); i++)
				map.put(nums[i], i);
			for(int i = 0; i < nums.size(); i++){
					int complement = target - nums[i];
					if (map.containsKey(complement) && map.get(complement) != i) 	
			    		return vector<int>{ i, map.get(complement) };
			}
		}          
	};
	

	