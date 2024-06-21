package com.l4yn3.microserviceseclab.controller;

import com.alibaba.fastjson.JSON;
import com.l4yn3.microserviceseclab.data.Teacher;
import org.springframework.web.bind.annotation.*;

import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;

@RestController
@RequestMapping(value = "/fastjson")
public class FastJsonController {

    @PostMapping(value = "/create")
    public Teacher createActivity(@RequestBody String applyData,
                                  HttpServletRequest request, HttpServletResponse response){
        Teacher teachVO = JSON.parseObject(applyData, Teacher.class);
        System.out.println(teachVO);
        return teachVO;
    }
    // 快速排序
    @PostMapping(value = "/sort")
    public void sort(@RequestBody String applyData, HttpServletResponse response){
        Teacher[] teachers = JSON.parseArray(applyData, Teacher.class).toArray(new Teacher[0]);
        for (int i=0;i<teachers.length-1;i++){
            for (int j=i+1;j<teachers.length;j++){
                if (teachers[i].getAge() > teachers[j].getAge()){
                    Teacher temp = teachers[i];
                    teachers[i] = teachers[j];
                    teachers[j] = temp;
                }
            }
        }
        System.out.println("排序后：" + JSON.toJSONString(teachers));
    }
}

// 测试333