import {useState,useEffect,useRef} from 'react';

const useTimer = () => {
const [sec,setSec]=useState(0);
const [min,setMin]=useState(0);
const [hrs,setHrs]=useState(0);
const [timerOn,setTimerOn]=useState(false);

const timerRef=useRef(null);
const handleTimer=()=>{
   if(timerOn){
    setTimerOn(false);
    setHrs(0);
    setMin(0);
    setSec(0);
     return clearInterval(timerRef.current);
   }
   setTimerOn(true);
}
useEffect(()=>{
    if(timerOn){
    timerRef.current=setInterval(() => {
    if(min===59 && sec===59){
        setMin(0);
        setSec(0);
        setHrs(hrs+1);
    }
    else if(sec===59){
        setMin(min+1);
        setSec(0);

    }else{
        setSec(sec+1);
    }
    },1000);
}
    return ()=>clearInterval(timerRef.current);
})
  return {sec,min,hrs,timerOn,handleTimer}
}

export default useTimer;